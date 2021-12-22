package server

import (
	"container/list"
	"context"
	b64 "encoding/base64"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog/log"
)

// Represents an operation received from client
type Op struct {
	Name  string
	Key   string
	Value string
}

//Returns Name, base64 encoded key and value
// separated by comma
func (op *Op) Marshal() string {
	return strings.Join([]string{
		op.Name,
		b64.StdEncoding.EncodeToString([]byte(op.Key)),
		b64.StdEncoding.EncodeToString([]byte(op.Value)),
	}, ",")
}

func (op *Op) Unmarshal(s string) error {
	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		return errors.New("Could not parse input.")
	}
	op.Name = string(parts[0])
	if len(parts) > 1 {
		k, _ := b64.StdEncoding.DecodeString(parts[1])
		op.Key = string(k)
	}
	if len(parts) > 2 {
		v, _ := b64.StdEncoding.DecodeString(parts[2])
		op.Value = string(v)
	}
	return nil
}

// Item to store in memory
type Item struct {
	Key   string
	Value string
	el    *list.Element
}

func NewItem(key, value string) *Item {
	return &Item{Key: key, Value: value}
}

//Ordered map, stores inserting order
type Storage struct {
	items map[string]*Item
	list  *list.List
}

func NewStorage() *Storage {
	return &Storage{
		items: make(map[string]*Item),
		list:  list.New(),
	}
}

func (s *Storage) AddItem(item *Item) error {
	log.Debug().Msgf("Adding item %#v", item)
	if item.Key == "" {
		return errors.New("Key could not be empty string.")
	}
	item.el = s.list.PushBack(item)
	s.items[item.Key] = item
	return nil
}

func (s *Storage) GetItem(key string) (*Item, bool) {
	i, e := s.items[key]
	return i, e
}

//Returns true if item was deleted, false otherwise
func (s *Storage) RemoveItem(key string) bool {
	item, exists := s.items[key]
	if !exists {
		return false
	}
	s.list.Remove(item.el)
	delete(s.items, key)
	return true
}

func el2item(el *list.Element) *Item {
	if el == nil {
		return nil
	}
	return el.Value.(*Item)
}

func (s *Storage) Oldest() *Item {
	return el2item(s.list.Front())
}

func (item *Item) Next() *Item {
	return el2item(item.el.Next())
}

//Returns list of items, oldest goes first
func (s *Storage) GetAllItems() []*Item {
	r := []*Item{}
	for item := s.Oldest(); item != nil; item = item.Next() {
		r = append(r, item)
	}
	return r
}

//Runs inside a go routine and process every sqs message received from
// inCh, puts an output to outCh
func (s *Storage) processMsg(svc *sqs.SQS, queue string, inCh chan *sqs.Message, outCh chan string) {
	for {
		msg := <-inCh
		log.Debug().Msgf("Processing %#v", msg)
		op := &Op{}
		err := op.Unmarshal(*msg.Body)
		if err != nil {
			log.Error().Msg("Could not unmarshal body.")
			op.Name = "foo" // so it is ignored in switch
		}
		log.Debug().Msgf("Operation is: %#v", op)
		switch op.Name {
		case "AddItem":
			err := s.AddItem(NewItem(op.Key, op.Value))
			if err != nil {
				log.Error().Msg(err.Error())
			}
			log.Debug().Msgf("Item with key %s has added.", op.Key)
		case "RemoveItem":
			if s.RemoveItem(op.Key) {
				log.Debug().Msgf("Item with key %s removed", op.Key)
			}
		case "GetItem":
			item, exists := s.GetItem(op.Key)
			if exists {
				log.Debug().Msgf("Sending to file %s", item.Value)
				outCh <- item.Value
			}
		case "GetAllItems":
			for _, item := range s.GetAllItems() {
				log.Debug().Msgf("Sending to file %s", item.Value)
				outCh <- item.Value
			}
		default:
			log.Warn().Msgf("No handler for operation %#v", op)
		}
		_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queue),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			log.Error().Msg(err.Error())
		}

	}
}

// Loop for polling messages from sqs queue
func pollSQS(svc *sqs.SQS, queue string, chn chan *sqs.Message) {
	for {
		r, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            aws.String(queue),
			MaxNumberOfMessages: aws.Int64(1),
			VisibilityTimeout:   aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(1),
		})
		if err != nil {
			log.Error().Msg(err.Error())
		}
		if len(r.Messages) > 0 {
			for _, msg := range r.Messages {
				chn <- msg
			}
		}
	}

}

//outCh for putting strings into output file
func (s *Storage) Listen(ctx context.Context, svc *sqs.SQS, queue string, outCh chan string) {
	toProcess := make(chan *sqs.Message)
	go s.processMsg(svc, queue, toProcess, outCh)
	go pollSQS(svc, queue, toProcess)
	<-ctx.Done()
}
