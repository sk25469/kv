package main

// func BenchmarkPubSub(b *testing.B) {
// 	ps := models.NewPubSub()
// 	// ch := ps.Subscribe("topic")

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		ps.Publish("topic", "message")
// 		<-ch
// 	}
// }

// func BenchmarkPubSubParallel(b *testing.B) {
// 	ps := models.NewPubSub()
// 	ch := ps.Subscribe("topic")

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			ps.Publish("topic", "message")
// 			<-ch
// 		}
// 	})
// }

// func BenchmarkPubSubParallelWithLock(b *testing.B) {
// 	ps := models.NewPubSub()
// 	ch := ps.Subscribe("topic")

// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			ps.Publish("topic", "message")
// 			ps.Mutex.Lock()
// 			<-ch
// 			ps.Mutex.Unlock()
// 		}
// 	})
// }
