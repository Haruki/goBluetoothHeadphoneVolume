package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

var mmde *wca.IMMDeviceEnumerator

func main() {
	log.SetFlags(0)
	log.SetPrefix("error: ")

	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return err
	}

	defer ole.CoUninitialize()

	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}

	defer mmde.Release()

	callback := wca.IMMNotificationClientCallback{
		OnDefaultDeviceChanged: onDefaultDeviceChanged,
		OnDeviceAdded:          onDeviceAdded,
		OnDeviceRemoved:        onDeviceRemoved,
		OnDeviceStateChanged:   onDeviceStateChanged,
		OnPropertyValueChanged: onPropertyValueChanged,
	}

	mmnc := wca.NewIMMNotificationClient(callback)

	if err := mmde.RegisterEndpointNotificationCallback(mmnc); err != nil {
		return err
	}

	select {
	case <-quit:
		fmt.Println("Received keyboard interrupt.")
		/*
			case <-time.After(5 * time.Minute):
				fmt.Println("Received timeout signal")
		*/
	}
	fmt.Println("Done")

	return nil
}

func onDefaultDeviceChanged(flow wca.EDataFlow, role wca.ERole, pwstrDeviceId string) error {
	fmt.Printf("Called OnDefaultDeviceChanged\t(%v, %v, %q)\n", flow, role, pwstrDeviceId)
	if role == 2 {
		var mmd *wca.IMMDevice
		if err := mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
			return err
		}
		defer mmd.Release()

		var ps *wca.IPropertyStore
		if err := mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
			return err
		}
		defer ps.Release()

		var pv wca.PROPVARIANT
		if err := ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
			return err
		}
		var deviceName string = pv.String()
		fmt.Printf("%s\n", deviceName)
		if strings.Contains(strings.ToLower(deviceName), "shanling") {
			fmt.Printf("Headphones detected: Setting Volume MAX!\n")
			var aev *wca.IAudioEndpointVolume
			if err := mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &aev); err != nil {
				return err
			}
			defer aev.Release()

			var volume float32 = 1.0
			if err := aev.SetMasterVolumeLevelScalar(volume, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func onDeviceAdded(pwstrDeviceId string) error {
	fmt.Printf("Called OnDeviceAdded\t(%q)\n", pwstrDeviceId)

	return nil
}

func onDeviceRemoved(pwstrDeviceId string) error {
	fmt.Printf("Called OnDeviceRemoved\t(%q)\n", pwstrDeviceId)

	return nil
}

func onDeviceStateChanged(pwstrDeviceId string, dwNewState uint64) error {
	//fmt.Printf("Called OnDeviceStateChanged\t(%q, %v)\n", pwstrDeviceId, dwNewState)

	return nil
}

func onPropertyValueChanged(pwstrDeviceId string, key uint64) error {
	//fmt.Printf("Called OnPropertyValueChanged\t(%q, %v)\n", pwstrDeviceId, key)
	return nil
}
