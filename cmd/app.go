package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"encoding/json"

	"time"

	"github.com/SherClockHolmes/webpush-go"
)

const (
	publicKey  = "BB1k6GlIrrB0TiO6WFbQFYZkNIdfOBzQN4yHa0xGvrhhBdqFhoibHxD_rGdUsFcc0p3UFfYf8kS-peymtBUc6M4"
	privateKey = "6KFXNWgb7dJaYfy4uAc5Zp158ghwc_Nx-XhHwXVU6GI"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/save", saveSubscription)
	mux.HandleFunc("/send", sendMessage)

	listener, err := net.Listen("tcp4", "0.0.0.0:6601")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("start")

	http.Serve(listener, mux)
}

func saveSubscription(resp http.ResponseWriter, req *http.Request) {
	log.Println("save subscription")

	resp.Header().Set("Access-Control-Allow-Origin", "http://localhost:2015")
	resp.Header().Set("Access-Control-Allow-Methods", "POST")
	resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if req.Method != "POST" {
		resp.WriteHeader(200)
		return
	}

	body := make([]byte, req.ContentLength)
	if _, err := req.Body.Read(body); err != nil && err != io.EOF {
		resp.WriteHeader(404)
		log.Println("read body error")
		return
	}

	go func() {
		if err := pushHello([]byte(body)); err != nil {
			log.Println("push hello error")
			log.Println(err)
		}
	}()

	if err := ioutil.WriteFile("subscription", body, 0644); err != nil {
		resp.WriteHeader(404)
		log.Println("write subscription to file error")
		log.Println(err)
		return
	}

	log.Printf("Subscription:\n%s\n", body)

	resp.WriteHeader(200)
}

func sendMessage(resp http.ResponseWriter, req *http.Request) {
	log.Println("send message")

	var err error

	data, err := ioutil.ReadFile("subscription")
	if err != nil && err != io.EOF {
		resp.WriteHeader(404)
		log.Println("read subscription from file error")
		log.Println(err)
		return
	}

	s := webpush.Subscription{}
	if err = json.Unmarshal(data, &s); err != nil && err != io.EOF {
		resp.WriteHeader(404)
		log.Println("unmarshal subscription error")
		log.Println(err)
		return
	}

	message := make([]byte, req.ContentLength)
	if _, err = req.Body.Read(message); err != nil && err != io.EOF {
		resp.WriteHeader(404)
		log.Println("read body error")
		log.Println(err)
		return
	}

	log.Printf("push to: %s", string(data))
	r, err := webpush.SendNotification([]byte(message), &s, &webpush.Options{
		Subscriber:      "example@example.com",
		TTL:             30,
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
	})
	if err != nil {
		resp.WriteHeader(404)
		log.Println("send nofication error")
		log.Println(err)
		return
	}
	defer r.Body.Close()

	resp.WriteHeader(200)
}

func pushHello(subscriptionData []byte) error {
	var err error

	log.Printf("push to: %s", string(subscriptionData))

	s := webpush.Subscription{}
	if err = json.Unmarshal(subscriptionData, &s); err != nil {
		return err
	}

	o := webpush.Options{
		TTL:             5,
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
	}

	r1, err := webpush.SendNotification([]byte("hello, pushes!"), &s, &o)
	if err != nil {
		return err
	}
	defer r1.Body.Close()

	time.Sleep(2 * time.Second)
	r2, err := webpush.SendNotification([]byte("more pushes!"), &s, &o)
	if err != nil {
		return err
	}
	defer r2.Body.Close()

	time.Sleep(2 * time.Second)
	r3, err := webpush.SendNotification([]byte("VERY MACH PUSHES!!!"), &s, &o)
	if err != nil {
		return err
	}
	defer r3.Body.Close()

	return nil
}
