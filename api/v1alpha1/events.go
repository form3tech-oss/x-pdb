/*
Copyright 2024 Form3.

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

package v1alpha1

// XPDBEventReason is a Reason that is attached to a
// Kind=XPDB event.
type XPDBEventReason string

const (
	// XPDBEventReasonInvalidConfiguration represents a invalid configuration,
	// such as a pod matched by multiple PDBs.
	XPDBEventReasonInvalidConfiguration XPDBEventReason = "InvalidConfiguration"
	// XPDBEventReasonBlocked represents a blocked disruption.
	XPDBEventReasonBlocked XPDBEventReason = "Blocked"
	// XPDBEventReasonAccepted represents a accepted disruption.
	XPDBEventReasonAccepted XPDBEventReason = "Accepted"
)
