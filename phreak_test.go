package main

import (
	"testing"
)

func Test_launch(t *testing.T) {
	ph := mkphreak()

	expectA := 0
	expectB := 1

	reg := &registration{}
	reg.newid = make(chan int)

	go ph.launch(reg)

	actualA := <-reg.newid

	if actualA != expectA {
		t.Error("Expected %v but received %v", expectA, actualA)
	}

	go ph.launch(reg)

	actualB := <-reg.newid

	if actualB != expectB {
		t.Error("Expected %v but received %v", expectB, actualB)
	}
}
func Test_ping(t *testing.T) {
	ph := mkphreak()

	reg := &registration{}
	reg.newid = make(chan int)

	go ph.launch(reg)

	rset := <-reg.newid

	res := &result{}
	res.port = 80
	res.set = uint64(rset)
	res.done = make(chan struct{})

	go ph.ping(res)

	<-res.done
}
func Test_badports(t *testing.T) {
	ph := mkphreak()

	reg := &registration{}
	reg.newid = make(chan int)

	go ph.launch(reg)

	rset := <-reg.newid

	res := &result{}
	res.port = 80
	res.set = uint64(rset)
	res.done = make(chan struct{})

	go ph.ping(res)

	<-res.done

	q := &query{}
	q.rset = uint64(rset)
	q.failports = make(chan []int)

	go ph.badports(q)

	expect := []int{90}

	actual := <-q.failports

	for i, acp := range actual {
		exp := expect[i]

		if acp != exp {
			t.Error("Expected %v but received %v", exp, acp)
		}
	}
}
