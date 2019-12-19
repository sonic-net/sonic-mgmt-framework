package transformer

import (
	"strings"

	// Go 1.11 doesn't let us download using the canonical import path
	// "github.com/godbus/dbus/v5"
	// This works around that problem by using gopkg.in instead
	"gopkg.in/godbus/dbus.v5"
)

// hostResult contains the body of the response and the error if any, when the 
// endpoint finishes servicing the D-Bus request.
type hostResult struct {
	Body	[]interface{}
	Err		error
}

// hostQuery calls the corresponding D-Bus endpoint on the host and returns
// any error and response body
func hostQuery(endpoint string, args ...interface{}) (result hostResult) {
	result_ch, err := hostQueryAsync(endpoint, args...)

	if err != nil {
		result.Err = err
		return
	}

	result = <-result_ch
	return
}

// hostQueryAsync calls the corresponding D-Bus endpoint on the host and returns
// a channel for the result, and any error
func hostQueryAsync(endpoint string, args ...interface{}) (chan hostResult, error) {
	var result_ch = make(chan hostResult, 1)
	conn, err := dbus.SystemBus()
	if err != nil {
		return result_ch, err
	}

	service := strings.SplitN(endpoint, ".", 2)
	const bus_name_base = "org.SONiC.HostService."
	bus_name := bus_name_base + service[0]
	bus_path := dbus.ObjectPath("/org/SONiC/HostService/" + service[0])

	obj := conn.Object(bus_name, bus_path)
	dest := bus_name_base + endpoint
	dbus_ch := make(chan *dbus.Call, 1)

	go func() {
		var result hostResult

		// Wait for a read on the channel
		call := <-dbus_ch

		if call.Err != nil {
			result.Err = call.Err
		} else {
			result.Body = call.Body
		}

		// Write the result to the channel
		result_ch <- result
	}()

	call := obj.Go(dest, 0, dbus_ch, args...)

	if call.Err != nil {
		return result_ch, call.Err
	}

	return result_ch, nil
}

// Example
/*
ztpAction calls the ZTP endpoint on the host and returns the status
func ztpAction(action string) (string, error) {
	var output string
	// result.Body is of type []interface{}, since any data may be returned by
	// the host server. The application is responsible for performing
	// type assertions to get the correct data.
	result := hostQuery("ztp." + action)

	if result.Err != nil {
		return output, result.Err
	}

	if action == "status" {
		// ztp.status returns an exit code and the stdout of the command
		// We only care about the stdout (which is at [1] in the slice)
		output, _ = result.Body[1].(string)
	}

	return output, nil
}

// The following uses the hostQueryAsync option
func ztpAction(action string) (string, error) {
	var output string
	// body is of type []interface{}, since any data may be returned by
	// the host server. The application is responsible for performing
	// type assertions to get the correct data.
	ch, err := hostQueryAsync("ztp." + action)

	if err != nil {
		return output, err
	}

	// Wait for the call to finish
	result := <-ch
	if result.Err != nil {
		return output, result.Err
	}

	if action == "status" {
		// ztp.status returns an exit code and the stdout of the command
		// We only care about the stdout (which is at [1] in the slice)
		output, _ = result.Body[1].(string)
	}

	return output, nil
}

*/

// showtechAction calls the Showtech endpoint on the host and returns
// the status
func showtechAction(date string) (string, error) {
	var output string
	// result.Body is of type []interface{}, since any data may be returned by
	// the host server. The application is responsible for performing
	// type assertions to get the correct data.
	result := hostQuery("Showtech.showtech_info", date)

	if result.Err != nil {
		return output, result.Err
	}

    // Showtech.showtech_info returns an exit code and the stdout of the command
    // We only care about the stdout (which is at offset 1 in the "Body".)
    output, _ = result.Body[1].(string)

	return output, nil
}
