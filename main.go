// main.go
package main

import (
    "marmur_excel_tool/ui"
	
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
)

func main() {
    
    a := app.NewWithID("com.marmur.exceltool")
    w := a.NewWindow("Marmur Excel Tool")
    
    ui.BuildUI(w)
    
    w.Resize(fyne.NewSize(900, 600))
    w.ShowAndRun()
}