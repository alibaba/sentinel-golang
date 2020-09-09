// Package api provides the topmost fundamental APIs for users using sentinel-golang.
// Users must initialize Sentinel before loading Sentinel rules. Sentinel support three ways to perform initialization:
//
//  1. api.InitDefault(), using default config to initialize.
//  2. api.InitWithConfig(confEntity *config.Entity), using customized config Entity to initialize.
//  3. api.InitWithConfigFile(configPath string), using yaml file to initialize.
//
// Here is the example code to use Sentinel:
//
//  import sentinel "github.com/alibaba/sentinel-golang/api"
//
//  err := sentinel.InitDefault()
//  if err != nil {
//      log.Fatal(err)
//  }
//
//  //Load sentinel rules
//  _, err = flow.LoadRules([]*flow.Rule{
//      {
//          Resource:        "some-test",
//          MetricType:      flow.QPS,
//          Count:           10,
//          ControlBehavior: flow.Reject,
//      },
//  })
//  if err != nil {
//      log.Fatalf("Unexpected error: %+v", err)
//      return
//  }
//  ch := make(chan struct{})
//  for i := 0; i < 10; i++ {
//      go func() {
//          for {
//              e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
//              if b != nil {
//                  // Blocked. We could get the block reason from the BlockError.
//                  time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
//              } else {
//                  // Passed, wrap the logic here.
//                  fmt.Println(util.CurrentTimeMillis(), "passed")
//                  time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
//                  // Be sure the entry is exited finally.
//                  e.Exit()
//              }
//          }
//      }()
//  }
//  <-ch
//
package api
