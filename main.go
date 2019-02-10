package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"unicode"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

// LambdaHandler for mock aws lambda handler
type LambdaHandler struct {
	handle func(ctx context.Context, b []byte) ([]byte, error)
}

// Invoke ...
func (h LambdaHandler) Invoke(ctx context.Context, b []byte) ([]byte, error) {
	return h.handle(ctx, b)
}

func wrap(handler interface{}) LambdaHandler {
	typ := reflect.TypeOf(handler)
	if typ.Kind() != reflect.Func {
		panic(fmt.Errorf("handler kind %s is not %s", typ.Kind(), reflect.Func))
	}
	if typ.NumIn() != 2 {
		panic(fmt.Errorf("handler args %d is not 2", typ.NumIn()))
	}

	args := reflect.New(typ.In(1))

	return LambdaHandler{func(ctx context.Context, b []byte) ([]byte, error) {
		var requuid string
		if lc, ok := lambdacontext.FromContext(ctx); ok {
			requuid = lc.AwsRequestID
		}

		log.Printf("%s[%s] with uuid: %s, args: %v", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid, string(b))

		var req events.APIGatewayProxyRequest
		if err := json.Unmarshal(b, &req); err != nil {
			log.Printf("%s[%s] with uuid: %s, unmarshal failed, err: %v", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid, err)
			return nil, err
		}

		args := reflect.Indirect(args)
		for i := 0; i < args.NumField(); i++ {
			key, ok := args.Type().Field(i).Tag.Lookup("json")

			// take filed name as key
			// lower case the first rune
			if !ok {
				s := args.Type().Field(i).Name
				for i, rn := range s {
					key = string(unicode.ToLower(rn)) + s[i+1:]
					break
				}
			}

			if val, ok := req.PathParameters[key]; ok {
				args.Field(i).Set(reflect.ValueOf(val))
				continue
			}

			if val, ok := req.QueryStringParameters[key]; ok {
				args.Field(i).Set(reflect.ValueOf(val))
				continue
			}
		}

		log.Printf("%s[%s] with uuid: %s, exec handler with args: %+v", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid, args)

		// exec handler
		var resp []reflect.Value
		if err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recover: %v", r)
				}
			}()
			resp = reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(ctx), args})
			return nil
		}(); err != nil {
			log.Printf("%s[%s] with uuid: %s, exec handler failed, err: %v", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid, err)
			return nil, err
		}

		if len(resp) > 0 {
			if i := resp[len(resp)-1].Interface(); i != nil {
				err, ok := i.(error)
				if !ok {
					panic(fmt.Errorf("the last of returns must be error, got %+v", i))
				}
				if err != nil {
					log.Printf("%s[%s] with uuid: %s, handler returns err, err: %v", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid, err)
					return nil, err
				}
			}
		}

		if len(resp) > 1 {
			b, err := json.Marshal(resp[0].Interface())
			if err != nil {
				return nil, err
			}
			log.Printf("%s[%s] with uuid: %s, handler returns: %v", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid, string(b))
			return b, nil
		}

		log.Printf("%s[%s] with uuid: %s, handler finished with empty result", lambdacontext.FunctionName, lambdacontext.FunctionVersion, requuid)
		return nil, nil
	}}
}

// main
func main() {
	lambda.StartHandler(wrap(handler))
}

type request struct {
	Name string `json:"name"`
}

type response struct {
	Name     string
	Greeting string
}

func handler(_ context.Context, req request) (events.APIGatewayProxyResponse, error) {
	log.Printf("handler with %+v", req)
	if len(req.Name) == 0 {
		return events.APIGatewayProxyResponse{}, errors.New("name, please")
	}

	b, err := json.Marshal(&response{Name: req.Name, Greeting: "hello"})
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	resp := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(b),
	}
	return resp, nil
}
