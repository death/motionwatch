package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/death/motionwatch/detector"
	"github.com/lazywei/go-opencv/opencv"
	"github.com/scorredoira/email"
)

var (
	initialSleep   = flag.Duration("initial", time.Duration(10*time.Minute), "Duration to sleep before detecting motion.")
	interval       = flag.Duration("interval", time.Duration(500*time.Millisecond), "Duration between frame grabs.")
	minDev         = flag.Float64("mindev", 20.0, "Minimum std deviation for motion detection.")
	notifyInterval = flag.Duration("notify", time.Duration(1*time.Hour), "Min. duration between notifications.")
	user           = flag.String("user", "", "Gmail user name")
	password       = flag.String("password", "", "Gmail user password")

	lastNotified = time.Time{}
)

func main() {
	flag.Parse()

	if *user == "" {
		log.Fatal("Need gmail user name.")
	}

	if *password == "" {
		log.Fatal("Need gmail user password.")
	}

	time.Sleep(*initialSleep)

	cap := opencv.NewCameraCapture(0)
	if cap == nil {
		log.Fatal("Can't open camera.")
	}
	defer cap.Release()

	det := detector.New(&detector.Params{
		DevThreshold:   *minDev,
		PhaseThreshold: 3,
		QueryFrame:     cap.QueryFrame,
	})
	defer det.Close()

	log.Println("Watching for motion")
	for {
		if im := det.Detect(); im != nil {
			detectedMotion(im)
		}
		time.Sleep(*interval)
	}
}

func detectedMotion(im *opencv.IplImage) {
	t := time.Now()
	d := t.Sub(lastNotified)
	if d < *notifyInterval {
		return
	}
	lastNotified = t

	log.Printf("Motion detected!")

	const tempFile = "/tmp/im.png"
	opencv.SaveImage(tempFile, im, 0)
	m := email.NewMessage("Motion detected", fmt.Sprintf("Motion detected at %v\n", t))
	m.From = *user + "@gmail.com"
	m.To = []string{*user + "@gmail.com"}
	err := m.Attach(tempFile)
	if err != nil {
		log.Printf("Couldn't attach image: %v\n", err)
	}
	err = email.Send("smtp.gmail.com:587",
		smtp.PlainAuth("", *user, *password, "smtp.gmail.com"),
		m)
	if err != nil {
		log.Printf("Couldn't send mail: %v\n", err)
		return
	}

	log.Printf("Sent email...\n")
}
