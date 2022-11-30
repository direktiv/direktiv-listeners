package main

import (
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"

	"github.com/scorredoira/email"
)

var (
// mailServer = "192.168.0.177:2525"
)

func main() {

	m := email.NewMessage("test subject", "body")
	m.From = mail.Address{
		Name:    "Not Me",
		Address: "who@am.i",
	}

	server := os.Args[1]
	fmt.Printf("sending with %s\n", server)

	files, err := ioutil.ReadDir("./attachments")
	if err != nil {
		panic(err)
	}

	for i := range files {
		f := files[i]
		fmt.Printf("attaching %v\n", f.Name())
		err = m.Attach(fmt.Sprintf("./attachments/%s", f.Name()))
		if err != nil {
			panic(err)
		}
	}

	to := []string{}
	for a := range os.Args {
		if a > 1 {
			t := os.Args[a]
			to = append(to, t)
		}
	}
	m.To = to

	fmt.Printf("sending to %s\n", to)

	// fmt.Printf("%s\n", string(m.Bytes()))

	err = email.Send(server, nil, m)
	if err != nil {
		fmt.Printf("can not send email: %v", err)
	}

}
