# A demo for aws lambda and ApiGateway integration.

### Something unusual

In `main()`, use `lambda.StartHandler(...)` and a custom `wrap()`
~~~go
func main() {
	lambda.StartHandler(wrap(handler))
}
~~~

Your `wrap()` is something like this:
~~~go
func wrap(handler interface{}) LambdaHandler {

	.....

	return LambdaHandler{func(ctx context.Context, b []byte) ([]byte, error) {

		......

		// the before func
		aspect.before()

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
			return nil, err
		}

		// below are aws standard output
		// you can put some after func here, before or after these
		if len(resp) > 0 {
			if i := resp[len(resp)-1].Interface(); i != nil {
				err, ok := i.(error)
				if !ok {
					panic(fmt.Errorf("the last of returns must be error, got %+v", i))
				}
				if err != nil {
					return nil, err
				}
			}
		}

		if len(resp) > 1 {
			b, err := json.Marshal(resp[0].Interface())
			if err != nil {
				return nil, err
			}
			return b, nil
		}

		return nil, nil
	}}
}
~~~

### For local test:

```
$ make local
```

then:
```
$ curl -i http://127.0.0.1:3000/greeting/foo
```

or you may want to deploy it to aws cloundformation:
change the `STACK-NAME` and aws s3 path to your own.
```
make deploy
```