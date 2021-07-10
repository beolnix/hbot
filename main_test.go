package main

import (
	"testing"
)

func TestRatio(t *testing.T) {
	var tests = []struct {
		name   string
		status Status
	}{
		{
			name: "positive",
			status: Status{
				Login: "positive",
				BlameStatus: BlameStatus{
					Sent:     2,
					Received: 4,
				},
			},
		}, {
			name: "negative",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     4,
					Received: 2,
				},
			},
		}, {
			name: "0 received",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     4,
					Received: 0,
				},
			},
		}, {
			name: "0 sent",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     0,
					Received: 4,
				},
			},
		}, {
			name: "0 received",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     8,
					Received: 0,
				},
			},
		}, {
			name: "0 sent",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     0,
					Received: 10,
				},
			},
		}, {
			name: "0 sent",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     40,
					Received: 0,
				},
			},
		}, {
			name: "beolnix",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     5,
					Received: 2,
				},
			},
		}, {
			name: "first",
			status: Status{
				Login: "negative",
				BlameStatus: BlameStatus{
					Sent:     1,
					Received: 0,
				},
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("for %s sent: %v, receive: %v - %s", tt.name, tt.status.BlameStatus.Sent, tt.status.BlameStatus.Received, prettyPrintStatus(tt.status))
		})
	}
}
