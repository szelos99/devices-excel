package main


func main() {
	ui := NewUi()
	ui.window.Show()
	ui.app.Run()
	ui.CloseUI()
}

