package server

import (
	"testing"

        "github.com/johnny-morrice/sensephreak/util"
)

func Test_launch(t *testing.T) {
	ph := testphreak()

	expectA := regisreply{}
	expectA.userid = 0
	expectA.scanid = 0
	expectB := regisreply{}
	expectB.userid = 1
	expectB.scanid = 1

	reg := registration{}
	reg.userid = nouser
	reg.reply = make(chan regisreply)

	go ph.launch(reg)

	actualA := <-reg.reply

	if actualA != expectA {
		t.Error("Expected %v but received %v", expectA, actualA)
	}

	go ph.launch(reg)

	actualB := <-reg.reply

	if actualB != expectB {
		t.Error("Expected %v but received %v", expectB, actualB)
	}
}
func Test_ping(t *testing.T) {
	ph := testphreak()

	reg := registration{}
	reg.userid = nouser
	reg.reply = make(chan regisreply)

	go ph.launch(reg)

	reply := <-reg.reply

	if reply.err != nil {
		t.Error(reply.err)
	}

	ping := ping{}
	ping.port = 80
	ping.set = uint64(reply.scanid)
	ping.reply = make(chan pingreply)

	go ph.ping(ping)

	pingreply := <-ping.reply

	if pingreply.err != nil {
		t.Error(pingreply.err)
	}
}
func Test_badports(t *testing.T) {
	ph := testphreak()

	reg := registration{}
	reg.userid = nouser
	reg.reply = make(chan regisreply)
        reg.StartPort = 80
        reg.EndPort = 82

	go ph.launch(reg)

	reply := <-reg.reply

	if reply.err != nil {
		t.Error(reply.err)
	}

	ping := ping{}
	ping.port = 80
	ping.set = uint64(reply.scanid)
	ping.reply = make(chan pingreply)

	go ph.ping(ping)

	pingreply := <-ping.reply

	if pingreply.err != nil {
		t.Error(pingreply.err)
	}

	q := query{}
	q.rset = uint64(reply.scanid)
	q.reply = make(chan queryreply)

	go ph.badports(q)

	expect := []util.PortStatus{
                util.PortStatus{
                        Port: 80,
                        State: util.PortOk,
                },
                util.PortStatus{
                        Port: 81,
                        State: util.PortBlocked,
                },
                util.PortStatus{
                        Port: 82,
                        State: util.PortOmitted,
                },
        }

	qreply := <-q.reply

	if qreply.err != nil {
		t.Error(qreply.err)
	}

	for i, acp := range qreply.portinfo {
		exp := expect[i]

		if acp != exp {
			t.Error("Expected", exp, "but received", acp)
		}
	}
}

func testphreak() *phreak {
        s := Server{}

	ph := mkphreak(s)
        ph.tests = mktests()

        return ph
}
