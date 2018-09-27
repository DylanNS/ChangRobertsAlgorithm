package main
import (
"fmt"
"net"
"os"
"strconv"
"time"
"bufio"
"strings"
)
//Variáveis globais interessantes para o processo
var err string
var myPort string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
 //dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
 //mensagens dos outros processos)
var id int //numero identificador do processo
var participant bool //participando ou nao da eleicao
var leader int


func CheckError(err1 error){
	if err1 != nil {
		fmt.Println("Erro: ", err1)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func doServerJob() {
//Ler (uma vez somente) da conexão UDP a mensagem
//Escreve na tela a msg recebida


	 buf := make([]byte, 1024)
	 for {

		 n, _, err := ServConn.ReadFromUDP(buf)
		 msg_received := string (buf[0:n])
		 //pegar id da mensagem
		 slice_msg := msg_received[1:n]
		 msg_id, err := strconv.Atoi(slice_msg)
		 
		 fmt.Println("\nReceived mensage: ",msg_received)
		 if strings.HasPrefix(msg_received, "S"){

			//Compara msg_id com seu id
			reception_stage1(msg_id)

		 } else if strings.HasPrefix(msg_received, "F"){

			if msg_id != id {
				//Marks itself as non-participant
				participant = false
				//records the elected id
				leader = msg_id
				//Forwards the elected message
				next := id % nServers + 1
				fmt.Printf("\n Encaminhando mensagem: %s para P%d\n",msg_received,next)
				doClientJob(id,msg_received)
			} else {
				//Election is over
				fmt.Println("\nEvery other process knows the leader: ",leader,"\n")
			}

		 } else if strings.HasPrefix(msg_received,"M"){
			//inicia eleicao 
			initElection()
		 } else {
			fmt.Println("Error: mensagem estranha -> ",err,msg_received)
		 }

		 if err != nil {
		 	fmt.Println("Error: ",err)
		 } 
 	}
}
func doClientJob(otherProcess int,mymsg string) {
//Envia uma mensagem (com valor i) para o servidor do processo
//otherServer

     buf := []byte(mymsg)
     _,err := CliConn[otherProcess % nServers].Write(buf)
     if err != nil {
        fmt.Println(mymsg, err)
    }

}

//Inicia eleicao enviando S(id) para P(id+1)
func initElection(){

	fmt.Println("\nIniciando eleicao!\n")
	participant = true
	next := (id) % nServers + 1
	fmt.Printf("Enviado S%d para P%d  \n",id,next)
	//construi a mensagem Sid
	msgstream := "S" + strconv.Itoa(id)
	//envia para Pnext
	doClientJob(id,msgstream)
}

//send msg to Pnext
func forward_msg(msg int){

    //Construi a mensagem
	msgstream := string ("S") + strconv.Itoa(msg)
	//encaminha para o proximo
	next := (id) %nServers +1
	fmt.Printf("\nEncaminhando mensagem: %s para P%d\n",msgstream,next)
	doClientJob(id,msgstream)

}

func reception_stage1 (q int){

	if (q > id ){
		participant = true
		forward_msg(q) //send Sq to Pnext
	} else if q < id {
		if !participant{
			//Se nao tiver participando
			//inicia a sua candidatura
			participant = true
			forward_msg(id) //makes itself election
		}
		//senao apenas discarta a msg
	} else {
		// become leader
		leader = id
		//New leader elected
		fmt.Println("\nNew Leader elected: ",leader)
		//start stage 2
		fmt.Println("\nElection is over")
		fmt.Println("\nStarting the stage 2.")
		init_stage2()
	}
}

func init_stage2(){

	participant = false
	//construir msg
	msgstream := "F" + strconv.Itoa(leader)
	//encaminhar mensagem
	next := (id) %nServers +1
	fmt.Printf("\nEncaminhando mensagem: %s para P%d\n",msgstream,next)
	doClientJob(id,msgstream)
}


func initConnections() {
	id, _ = strconv.Atoi(os.Args[1])
	myPort = os.Args[ id + 1]
	nServers = len(os.Args) - 2
	/*Esse 2 tira o nome (no caso Process) e tira a primeira porta
	(que é a minha). As demais portas são dos outros processos*/
	//Outros códigos para deixar ok a conexão do meu servidor
	//Outros códigos para deixar ok as conexões com os servidores
	//dos outros processos
	connections := make([]*net.UDPConn, nServers, nServers)
	
	for i:=0; i<nServers; i++ {

		port := os.Args[i+2]
		
			ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1" + string (port) )
			PrintError(err)
 
    		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
 		   	PrintError(err)
 
    		connections[i], err = net.DialUDP("udp",LocalAddr, ServerAddr)
			PrintError(err)
		
	}
	CliConn = connections

	 /* Lets prepare a address at any address at port 10001*/   
	 ServerAddr,err := net.ResolveUDPAddr("udp", myPort)
	 CheckError(err)
	
	 /* Now listen at selected port */
	 ServConn, err = net.ListenUDP("udp", ServerAddr)
	 CheckError(err)

}

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal input
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}
	
func startMultiElection(){


	 //Construi a mensagem
	 msgstream := string ("M") + strconv.Itoa(id)
	 //encaminha para o proximo
	 next := (id) %nServers +1
	 anterior := (((id-2)%nServers)+nServers) % nServers
	 fmt.Printf("\nEncaminhando mensagem: %s para P%d\n",msgstream,next)
	 doClientJob(id,msgstream)
	 panterior := anterior + 1
	 fmt.Printf("\nEncaminhando mensagem: %s para P%d\n",msgstream,panterior)
	 
	 doClientJob( anterior, msgstream)
}

func main(){
	initConnections()
	//sem leader
	leader = -1
	//eleicao sem participantes
	participant = false
	//O fechamento de conexões devem ficar aqui, assim só fecha
	//conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}
	//Todo Process fará a mesma coisa: ouvir msg e mandar infinitos
	//i’s para os outros processos
	ch := make(chan string)
	go readInput(ch)
	for {
			
		//Server
		go doServerJob()
		// When there is a request (from stdin). Do it!
		select {
			case x, valid := <- ch :
				if valid {
                        
						if ( x == "start"){
							initElection() //inicia a eleicao. Envia Sid para Pnext
						} else if x == "multi" {

							startMultiElection()
						} else{

							if leader == -1{
								fmt.Printf("No leader\n")
							} else {
								fmt.Printf("Leader: %d \n",leader)
							}
							
						}
				} else {
						 
					fmt.Println("Channel closed!")
						
				}
				
			default:
			
				// Do nothing in the non-blocking approach.
			
				time.Sleep(time.Second * 1)
		}
			
		// Wait a while
		time.Sleep(time.Second * 1)
	}
}
		