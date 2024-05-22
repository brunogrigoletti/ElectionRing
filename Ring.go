package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // Tipo da mensagem para fazer o controle
	corpo [3]int // Conteúdo da mensagem (IDs)
}

var (
	chans = []chan mensagem{ // Vetor de canais para formar o anel de eleição
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// Mudar o processo 0 (canal de entrada 3) para falho

	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	fmt.Printf("Controle: confirmação %d\n", <-in)

	// Mudar o processo 1 (canal de entrada 0) para falho

	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// Matar outrOs processos com mensagens não conhecidas

	temp.tipo = 4
	chans[1] <- temp
	chans[2] <- temp

	fmt.Println("\nProcesso controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	var actualLeader int
	var bFailed bool = false // Todos inciam sem falha

	actualLeader = leader

	temp := <-in
	fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

	switch temp.tipo {
	case 2:
		{
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			controle <- -5
		}
	case 3:
		{
			bFailed = false
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			controle <- -5
		}
	default:
		{
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
			fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {
	wg.Add(5)

	go ElectionStage(0, chans[3], chans[0], 0) // Lider
	go ElectionStage(1, chans[0], chans[1], 0) // Processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // Processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // Processo 0

	fmt.Println("\nAnel de processos criado")

	go ElectionControler(controle)

	fmt.Println("\nProcesso controlador criado\n")

	wg.Wait()
}