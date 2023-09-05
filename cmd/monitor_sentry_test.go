package main

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestConvertStringToTags(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected []*tag
	}{
		{
			name:  "it should convert a list of string into tags",
			input: []string{"key1=value1", "key2=value2"},
			expected: []*tag{
				&tag{Key: "key1", Value: "value1"},
				&tag{Key: "key2", Value: "value2"},
			},
		},
		{
			name:     "it should not convert wrongly formatted tags",
			input:    []string{"key1"},
			expected: []*tag{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := convertStringToTags(test.input)
			if !reflect.DeepEqual(test.expected, output) {
				t.Errorf(
					"\ngiven %v\nexpected: %v\ngot: %v\n",
					spew.Sdump(test.input),
					spew.Sdump(test.expected),
					spew.Sdump(output),
				)
			}
		})
	}
}

type matchTagsInput struct {
	tagList      []*tag
	matchTagList []*tag
}

func TestMatchTags(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    matchTagsInput
		expected bool
	}{
		{
			name: "it should return true if the tags matches",
			input: matchTagsInput{
				tagList: []*tag{
					&tag{Key: "key2", Value: "value2"},
				},
				matchTagList: []*tag{
					&tag{Key: "key1", Value: "value1"},
					&tag{Key: "key2", Value: "value2"},
					&tag{Key: "key3", Value: "value3"},
				},
			},
			expected: true,
		},
		{
			name: "it should return false if not tags matches",
			input: matchTagsInput{
				tagList: []*tag{
					&tag{Key: "key2", Value: "value2"},
				},
				matchTagList: []*tag{
					&tag{Key: "key1", Value: "value1"},
					&tag{Key: "key3", Value: "value3"},
				},
			},
			expected: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := matchTags(test.input.tagList, test.input.matchTagList)
			if !reflect.DeepEqual(test.expected, output) {
				t.Errorf(
					"\ngiven %v\nexpected: %v\ngot: %v\n",
					spew.Sdump(test.input),
					spew.Sdump(test.expected),
					spew.Sdump(output),
				)
			}
		})
	}
}

type matchEventsInput struct {
	eventList []*sentryEvent
	message   string
	tagList   []*tag
	useRegexp bool
}

func TestMatchEvents(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    matchEventsInput
		expected []*sentryEvent
	}{
		{
			name: "it should match events by message and tags",
			input: matchEventsInput{
				message: "This is an event",
				tagList: []*tag{
					&tag{Key: "key2", Value: "value2"},
					&tag{Key: "key3", Value: "value3"},
				},
				eventList: []*sentryEvent{
					&sentryEvent{
						Message: "This is an event",
						Tags: []*tag{
							&tag{Key: "key1", Value: "value1"},
						},
					},
					&sentryEvent{
						Message: "This is an event",
						Tags: []*tag{
							&tag{Key: "key1", Value: "value1"},
							&tag{Key: "key2", Value: "value2"},
							&tag{Key: "key3", Value: "value3"},
						},
					},
					&sentryEvent{
						Message: "",
						Tags: []*tag{
							&tag{Key: "key2", Value: "value2"},
							&tag{Key: "key3", Value: "value3"},
						},
					},
					&sentryEvent{
						Message: "This is an event",
						Tags: []*tag{
							&tag{Key: "key2", Value: "value2"},
							&tag{Key: "key3", Value: "value3"},
						},
					},
				},
			},
			expected: []*sentryEvent{
				&sentryEvent{
					Message: "This is an event",
					Tags: []*tag{
						&tag{Key: "key1", Value: "value1"},
						&tag{Key: "key2", Value: "value2"},
						&tag{Key: "key3", Value: "value3"},
					},
				},
				&sentryEvent{
					Message: "This is an event",
					Tags: []*tag{
						&tag{Key: "key2", Value: "value2"},
						&tag{Key: "key3", Value: "value3"},
					},
				},
			},
		},
		{
			name: "it should match events by regular expression",
			input: matchEventsInput{
				eventList: []*sentryEvent{
					&sentryEvent{Message: "This is a first event"},
					&sentryEvent{Message: "This is a second event"},
					&sentryEvent{Message: "This is a third event"},
				},
				message:   "first|second",
				useRegexp: true,
			},
			expected: []*sentryEvent{
				&sentryEvent{Message: "This is a first event"},
				&sentryEvent{Message: "This is a second event"},
			},
		},
		{
			name: "it should match all events if message and tag list is not provided",
			input: matchEventsInput{
				eventList: []*sentryEvent{
					&sentryEvent{Message: "This is a first event"},
					&sentryEvent{Message: "This is a second event"},
					&sentryEvent{Message: "This is a third event"},
				},
				message:   "",
				useRegexp: false,
			},
			expected: []*sentryEvent{
				&sentryEvent{Message: "This is a first event"},
				&sentryEvent{Message: "This is a second event"},
				&sentryEvent{Message: "This is a third event"},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			output, err := matchEvents(
				test.input.eventList,
				test.input.message,
				test.input.tagList,
				test.input.useRegexp,
			)
			if !reflect.DeepEqual(test.expected, output) || err != nil {
				t.Errorf("\ngiven: \n%v\nexpected: \n%v\ngot: \n%v",
					spew.Sdump(test.input),
					spew.Sdump(test.expected),
					spew.Sdump(output),
				)
			}
		})
	}
}
