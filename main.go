package main

import (
	"os" // sistema operacional
	"os/exec"
	"time"

	"fyne.io/fyne/v2" //GUI
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty" // pseudoterminais
)

func main() {
	a := app.New() // App Fyne
	w := a.NewWindow("Kariuki")

	ui := widget.NewTextGrid() // Cria a nova grade de texto
	ui.SetText("Terminal Kariuki")

	c := exec.Command("/bin/bash") // cria um comando para executar o bash
	p, err := pty.Start(c)         // p é um pseudoterminal que permite a comunicação entre o programa e o terminal
	// p é um objeto que implementa Read e Write. Ele é o canal de comunicação entre o programa e o terminal
	// Ele permite que o programa envie comandos para o terminal e receba a saída do terminal

	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill() // mata o processo quando o programa termina

	p.Write([]byte("ls\r")) // envia o comando ls para o terminal. O \r é necessário para que o comando seja executado (simula o Enter)
	time.Sleep(time.Second) // espera um segundo para que o comando seja executado (em prod leitura reativa é a melhor forma)
	b := make([]byte, 1024) // Lê até 1024 bytes da saída do terminal
	_, err = p.Read(b)      // Vai ler os dados que o processo bash imprimiu no stdout. Leitura é feita diretamente do pseudoterminal
	if err != nil {
		fyne.LogError("Failed to read pty", err)
	}

	// s := fmt.Sprintf("read bytes from pty.\nContent:%s",  string(b))
	ui.SetText(string(b)) // A resposta do ls foi lida com sucesso e foi exibida na grade de texto

	w.SetContent(
		fyne.NewContainerWithLayout(
			layout.NewGridWrapLayout(fyne.NewSize(420, 200)),
			ui, // É o conteúdo do output do terminal
		),
	)
	w.ShowAndRun() // Exibe a janela e inicia o loop principal da interface gráfica
}
