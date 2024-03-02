package pub

import (
	"log"

	"github.com/nats-io/nats.go"
)

func _(count int, log *log.Logger, nc *nats.Conn) {
	/*
		const channel = "db-logs"

		log.Print("Publisher started")

		data, err := os.ReadFile("data/model.json")
		if err != nil {
			log.Fatal("Error: failed reading file: ", err)
		}

		var order domain.Order
		_ = json.Unmarshal(data, &order)

		sc, err := stan.Connect("dev", "order-producer", stan.NatsConn(nc),
			stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
				log.Fatal("Error: NATS connection lost, reason: ", reason)
			}))

		if err != nil {
			log.Fatal("Error: publisher failed connecting to cluster: ", err)
		}

		fo, err := os.Create("data/uids.txt")
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		for i := 0; i < ordersCount; i++ {
			uid := "" // utils.GenerateUID19v2()
			_, _ = fo.Write([]byte(uid + "\n"))

			order.OrderUid = uid
			order.DateCreated = time.Now()

			data, err = json.Marshal(order)
			if err != nil {
				log.Fatal("Error: failed marshal data: ", err)
			}

			if err := sc.Publish(channel, data); err != nil {
				log.Fatal("Error: failed publish message: ", err)
			}

			if (i % 100) == 0 {
				log.Printf("Messages count statistics: %d", i)
			}

			time.Sleep(time.Millisecond * 1)
		}

		log.Print("Publisher finished")

		errFlush := nc.Flush()
		if errFlush != nil {
			panic(errFlush)
		}

		errLast := nc.LastError()
		if errLast != nil {
			panic(errLast)
		}
	*/
}
