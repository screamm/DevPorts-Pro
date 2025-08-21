package main

import "fyne.io/fyne/v2"

// Skapa en enkel ikon med ASCII-art som resource
var appIcon = &fyne.StaticResource{
	StaticName:    "icon.png",
	StaticContent: []byte{}, // Tom för nu, men vi kan lägga till riktigt ikon senare
}

// ASCII-art ikon för terminal-look
const terminalIcon = `
   ████████
  ██      ██
  ██ >_   ██
  ██      ██
  ██  📊  ██
  ██      ██
   ████████
`