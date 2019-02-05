/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

// This file contains a collection of methods that can be used from go-restful to
// generate Swagger API documentation for its models. Please read this PR for more
// information on the implementation: https://github.com/emicklei/go-restful/pull/215
//
// TODOs are ignored from the parser (e.g. TODO(andronat):... || TODO:...) if and only if
// they are on one line! For multiple line or blocks that you want to ignore use ---.
// Any context after a --- is ignored.
//
// Those methods can be generated by using hack/update-generated-swagger-docs.sh

// AUTO-GENERATED FUNCTIONS START HERE. DO NOT EDIT.
var map_Event = map[string]string{
	"":                         "Event is a report of an event somewhere in the cluster. It generally denotes some state change in the system.",
	"eventTime":                "Required. Time when this Event was first observed.",
	"series":                   "Data about the Event series this event represents or nil if it's a singleton Event.",
	"reportingController":      "Name of the controller that emitted this Event, e.g. `kubernetes.io/kubelet`.",
	"reportingInstance":        "ID of the controller instance, e.g. `kubelet-xyzf`.",
	"action":                   "What action was taken/failed regarding to the regarding object.",
	"reason":                   "Why the action was taken.",
	"regarding":                "The object this Event is about. In most cases it's an Object reporting controller implements. E.g. ReplicaSetController implements ReplicaSets and this event is emitted because it acts on some changes in a ReplicaSet object.",
	"related":                  "Optional secondary object for more complex actions. E.g. when regarding object triggers a creation or deletion of related object.",
	"note":                     "Optional. A human-readable description of the status of this operation. Maximal length of the note is 1kB, but libraries should be prepared to handle values up to 64kB.",
	"type":                     "Type of this event (Normal, Warning), new types could be added in the future.",
	"deprecatedSource":         "Deprecated field assuring backward compatibility with core.v1 Event type",
	"deprecatedFirstTimestamp": "Deprecated field assuring backward compatibility with core.v1 Event type",
	"deprecatedLastTimestamp":  "Deprecated field assuring backward compatibility with core.v1 Event type",
	"deprecatedCount":          "Deprecated field assuring backward compatibility with core.v1 Event type",
}

func (Event) SwaggerDoc() map[string]string {
	return map_Event
}

var map_EventList = map[string]string{
	"":         "EventList is a list of Event objects.",
	"metadata": "Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata",
	"items":    "Items is a list of schema objects.",
}

func (EventList) SwaggerDoc() map[string]string {
	return map_EventList
}

var map_EventSeries = map[string]string{
	"":                 "EventSeries contain information on series of events, i.e. thing that was/is happening continuously for some time.",
	"count":            "Number of occurrences in this series up to the last heartbeat time",
	"lastObservedTime": "Time when last Event from the series was seen before last heartbeat.",
	"state":            "Information whether this series is ongoing or finished.",
}

func (EventSeries) SwaggerDoc() map[string]string {
	return map_EventSeries
}

// AUTO-GENERATED FUNCTIONS END HERE
