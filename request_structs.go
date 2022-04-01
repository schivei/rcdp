package main

import (
	json2 "encoding/json"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
)

// Inner Request
type innerRequest struct {
	*dataRequest
	// XPATH Selector
	Selector string `json:"selector" example:"id(\"btn_login\")" format:"string"`
}

// Refresh Request
type refreshRequest struct {
	*dataRequest
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
}

// CP cookie parameter object.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Network#type-CookieParam
type CP struct {
	Name     string `json:"name"`               // Cookie name.
	Value    string `json:"value"`              // Cookie value.
	URL      string `json:"url,omitempty"`      // The request-URI to associate with the setting of the cookie. This value can affect the default domain, path, source port, and source scheme values of the created cookie.
	Domain   string `json:"domain,omitempty"`   // Cookie domain.
	Path     string `json:"path,omitempty"`     // Cookie path.
	Secure   bool   `json:"secure,omitempty"`   // True if cookie is secure.
	HTTPOnly bool   `json:"httpOnly,omitempty"` // True if cookie is http-only.
	// CookieSameSite enum:
	// * Strict - strict
	// * Lax - lax
	// * None - none
	SameSite network.CookieSameSite `json:"sameSite,omitempty" enum:"Strict,Lax,None" swaggertype:"string" example:"Strict"`       // Cookie SameSite type.
	Expires  *cdp.TimeSinceEpoch    `json:"expires,omitempty" swaggertype:"string" example:"2022-03-29T00:00:00Z" format:"string"` // Cookie expiration date, session cookie if not set
	// CookiePriority enum:
	// * Low - low
	// * Medium - medium
	// * High - hight
	Priority  network.CookiePriority `json:"priority,omitempty" swaggertype:"string" example:"Low" enum:"Low,Medium,Hight"` // Cookie Priority.
	SameParty bool                   `json:"sameParty,omitempty"`                                                           // True if cookie is SameParty.
	// CookieSourceScheme enum:
	// * Unset
	// * NonSecure
	// * Secure
	SourceScheme network.CookieSourceScheme `json:"sourceScheme,omitempty" swaggertype:"string" example:"Unset" enum:"Unset,NonSecure,Secure"` // Cookie source scheme type.
	SourcePort   int64                      `json:"sourcePort,omitempty"`                                                                      // Cookie source port. Valid values are {-1, [1, 65535]}, -1 indicates an unspecified port. An unspecified port value allows protocol clients to emulate legacy cookie scope for the port. This is a temporary ability and it will be removed in the future.
	PartitionKey string                     `json:"partitionKey,omitempty"`                                                                    // Cookie partition key. The site of the top-level URL the browser was visiting at the start of the request to the endpoint that set the cookie. If not set, the cookie will be set as not partitioned.
}

func (cp *CP) toCp() (*network.CookieParam, error) {
	if cp == nil {
		return nil, nil
	}
	json, err := json2.Marshal(cp)
	c := &network.CookieParam{}
	if err != nil {
		err = json2.Unmarshal(json, &c)
		if err != nil {
			c = nil
		}
	}
	return c, err
}

// Navigate Request
type navigateRequest struct {
	*dataRequest
	// Url to navigate
	Url string `json:"url" example:"https://example.com" format:"string"`
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
	// Headers to inject
	Headers map[string]string `json:"headers,omitempty" example:"\{\"Authorization\":\"Bearer lnalsijosdf\"\}" format:"json"`
	// Cookies to inject
	Cookies []*CP `json:"cookies,omitempty" format:"json" example:"\[\{\"Name\":\"_gat\",\"Value\":\"1\",\"Domain\":\".domain.com\",\"Path\":\"/\",\"Expires\":\"2022-03-29T00:00:00\",\"Priority\":\"High\"\}\]"`
}

type typeRequest struct {
	*dataRequest
	// XPATH Selector
	Selector string `json:"selector" example:"id(\"btn_login\")" format:"string"`
	// Data content with special keys like ArrowDown (\u0301)
	Content string `json:"content" example:"test@test.com\\u0301" format:"string"`
	// Normalized is content withour special keys
	NormalizedContent string `json:"normalized_content" example:"test@test.com" format:"string"`
	// Force type
	Force bool `json:"force,omitempty" example:"false" format:"bool"`
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
}

// XY Position
type XY struct {
	// X coordinate
	X float64 `json:"x" example:"100" format:"float64"`
	Y float64 `json:"y" example:"100" format:"float64"`
}

// Mouse Request
type mouseRequest struct {
	*dataRequest
	// XPATH Selector
	Selector *string `json:"selector,omitempty" example:"id(\"btn_login\")" format:"string"`
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
	// Click at position
	Position *XY `json:"position"`
	// Mouse Button:
	// * left - Left Button
	// * middle - Wheel Button
	// * right - Right Button
	Button *input.MouseButton `json:"button" swaggertype:"string" enum:"left,middle,right" extensions:"x-nullable" `
	// Number of consecutive clicks
	Times int `json:"times" example:"1" format:"int"`
	// Event type:
	// * click - Click Event
	// * mouseover - Mouseover Event
	// * mouseout - Mouseout Event
	// * scroll - Scroll Event
	// * mousedown - Mousedown Event
	// * mouseup - Mouseup Event
	Event string `json:"event" example:"click" format:"string" enum:"click,mouseover,mouseout,scroll,mousedown,mouseup"`
}

// Select Request
type selectRequest struct {
	*dataRequest
	// XPATH Selector
	Selector string `json:"selector" example:"id(\"sel_options\")" format:"string"`
	// Event type:
	// * index - The option position
	// * text - The option text
	// * partialtext - The option that contains a text
	// * regex - The option text/value matchs
	// * value - The option value
	SelectBy string `json:"select_by" example:"value" format:"string" enum:"index,text,partialtext,regex,value"`
	// The data for select_by filter
	Data []string `json:"data" example:"[\"city\"]" formart:"array,string"`
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
}

type attrRequest struct {
	*dataRequest
	// XPATH Selector
	Selector string `json:"selector" example:"id(\"btn_login\")" format:"string"`
	// Attribute name
	Attribute string `json:"attribute" example:"name" format:"string"`
	// Content data
	Data *string `json:"data" example:"el_name" format:"string"`
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
	// Setter: if true set value to attribute, otherwise retrieve data from attribute
	Set bool `json:"set" example:"true" format:"bool"`
}

type jsRequest struct {
	*dataRequest
	// javascript to be executed
	Script string `json:"script" example:""`
	// Wait for XPATH visible
	WaitFor *string `json:"wait_for,omitempty" example:"//form/input[2]" format:"string"`
}

func (dr *innerRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}

func (dr *navigateRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}

func (dr *typeRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}

func (dr *selectRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}

func (dr *attrRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}

func (dr *jsRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}

func (dr *refreshRequest) GetTimeout() int64 {
	if dr.dataRequest == nil {
		dr.dataRequest = &dataRequest{}
	}
	return dr.dataRequest.GetTimeout()
}
