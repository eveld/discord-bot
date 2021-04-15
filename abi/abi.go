package abi

import "C"
import "github.com/nicholasjackson/wasp/go-abi"

//export send_channel_message
func SendMessage(channel abi.WasmString, content abi.WasmString) int32

//export delete_channel_message
func DeleteMessage(channel abi.WasmString, id abi.WasmString) int32
