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
	controle = make(chan int) // Canal externo para forçar falhas
	wg       sync.WaitGroup
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem
	// O: Encerrar programa
	// 1: Iniciar eleição
	// 2: Falhar processo
	// 3: Acordar processo

	temp.tipo = 2 // Mudar o processo 0 (canal de entrada 3) para falho
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)
	
	temp.tipo = 1 // O processo 1 (canal de entrada 0) deve iniciar uma eleição
	chans[0] <- temp
	fmt.Printf("Controle: processo 0 falhou. Iniciar eleição\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)

	temp.tipo = 3 // O processo 0 será re-ativado
	chans[3] <- temp
	fmt.Printf("Controle: acordou o processo 0\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)

	temp.tipo = 1 // O processo 1, atual líder, deve iniciar uma nova eleição
	chans[0] <- temp
	fmt.Printf("Controle: processo 0 voltou. Iniciar eleição\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)

	temp.tipo = 1 // O processo 0 inicia uma eleição sem que qualquer processo tenha falhado
	chans[3] <- temp
	fmt.Printf("Controle: processo 0 falhou. Iniciar eleição\n")
	fmt.Printf("Controle: confirmação %d\n", <-in)

	temp.tipo = 0 // Encerrar o script de testes
	chans[0] <- temp
	chans[1] <- temp
	chans[2] <- temp
	chans[3] <- temp
	fmt.Println("\nTodos os processos pararam. Programa encerrado")

	fmt.Println("\nProcesso controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	var actualLeader int
	var bFailed bool = false // Todos inciam sem falha
	var hardStop bool = false

	actualLeader = leader

	for !hardStop {
		temp := <-in
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])
		switch temp.tipo {
			case 0:
				{
					fmt.Println("\nPrograma encerrado")
					controle <- -5
					hardStop = true
				}
			case 1:
				{
					if bFailed {
						fmt.Printf("%2d: começou eleição\n", TaskId)
						// Devo criar um novo tipo para a confirmação da eleição?
						temp.tipo = 'ELEICAO'
						// Como eu incluo o ID do processo na mensagem?
						temp.tipo = 'ID'
						// Como comunicar o resultado?
						out <- temp
					}
				}
			case 2:
				{
					bFailed = true
					fmt.Printf("%2d: falhou %v\n", TaskId, bFailed)
					fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
					controle <- -5
				}
			case 3:
				{
					bFailed = false
					fmt.Printf("%2d: acordou\n", TaskId)
					fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
					controle <- -5
				}
			default:
				{
					fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
					fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				}
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {
	wg.Add(5)

	go ElectionStage(0, chans[3], chans[0], 0) // Líder
	go ElectionStage(1, chans[0], chans[1], 0) 
	go ElectionStage(2, chans[1], chans[2], 0) 
	go ElectionStage(3, chans[2], chans[3], 0)

	fmt.Println("\nAnel de processos criado")

	go ElectionControler(controle)

	fmt.Println("\nProcesso controlador criado\n")

	wg.Wait()
}