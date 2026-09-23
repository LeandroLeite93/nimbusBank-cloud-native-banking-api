# NimbusBank API

Laboratório educacional de Backend, DevOps e Cloud Computing em Go. Simulará operações bancárias para estudo; não é um banco real nem software pronto para produção.

## Etapa atual

Fase 1, segundo passo: um servidor HTTP local com `GET /health` e `GET /customers`, que retorna um cliente fictício fixo. Sem dependências externas ou recursos cloud. Cadastro de clientes, contas e operações financeiras serão adicionados progressivamente, após validação do aprendizado.

## Pré-requisitos e execução

- Go 1.27.1 ou superior (versão registrada no módulo).
- Git para versionamento e curl para verificar HTTP.
- Um editor, como VS Code.

Neste ambiente, Go foi instalado com `brew install go`. Para outro Mac, também existe o [instalador oficial](https://go.dev/doc/install); escolha apenas um método de instalação.

Na raiz do projeto:

```sh
go version
go run .
```

`go run .` compila os arquivos do pacote atual em um executável temporário e o executa. O terminal permanece ocupado pelo servidor. Encerre com `Ctrl+C`.

O servidor escuta em `127.0.0.1:8080`, somente no próprio computador. `localhost` é o nome usado para acessar a própria máquina. Nenhum recurso cloud é criado.

## Arquitetura e leitura do código

```text
curl → HTTP na porta 8080 → roteador → healthHandler → resposta JSON
```

Uma API é um contrato de comunicação entre programas. REST é um estilo arquitetural orientado a recursos: futuramente, `GET /customers` consultará clientes e `POST /customers` criará um cliente. Método, caminho, headers e corpo compõem a requisição; status, headers e corpo compõem a resposta. JSON é um formato de representação, não uma definição de REST.

| Arquivo | Função |
| --- | --- |
| `go.mod` | Nome do módulo (`nimbusbank`) e versão de Go; não temos dependências externas |
| `main.go` | Entrada do programa, configuração da rota e handler |
| `.gitignore` | Evita versionar arquivos de ambiente, binários e arquivos locais do macOS |

Leia `main.go` nesta ordem:

1. `package main` identifica um programa executável. `func main()` é seu ponto de entrada.
2. Os imports disponibilizam JSON, logs e HTTP da biblioteca padrão.
3. `http.NewServeMux()` cria o roteador. `HandleFunc` associa método e caminho ao handler.
4. `http.ListenAndServe` recebe conexões e encaminha requisições ao roteador. Se não conseguir iniciar, por exemplo porque a porta está ocupada, `log.Fatalf` exibe o erro e encerra com código diferente de zero. A mensagem “Iniciando” indica a tentativa de inicialização.
5. `healthHandler` recebe um `http.ResponseWriter`, usado para escrever a resposta, e um `*http.Request`, com dados da requisição. Esta rota não precisa consultar os dados de `r`.
6. A struct agrupa os dados da resposta. `Status` começa com maiúscula para ser exportado e acessível ao encoder; a tag `json:"status"` define o nome no JSON.
7. O header `Content-Type` é definido antes do corpo. `json.NewEncoder(w).Encode(response)` escreve o JSON, seguido de uma quebra de linha. A primeira escrita envia implicitamente `200 OK`.
8. Se a escrita falhar, registramos o erro. Como a resposta pode já ter começado a ser enviada, não tentamos escrever outra resposta por cima.

`http.ResponseWriter` é uma interface: define operações que o handler pode usar sem conhecer a implementação da conexão. Não precisamos criar interfaces próprias nesta etapa.

Escolhemos a biblioteca padrão para entender HTTP diretamente. Frameworks podem acrescentar convenções úteis, mas não são necessários para esta rota. Manter um pacote facilita acompanhar o fluxo; a separação em camadas ocorrerá quando existirem responsabilidades que justifiquem isso.

## Verificação manual

Com o servidor aberto, execute em outro terminal:

```sh
curl -i http://localhost:8080/health
```

Verifique `200 OK`, `Content-Type: application/json` e o corpo:

```json
{"status":"ok"}
```

`-i` mostra também os headers da resposta. Verifique os demais comportamentos:

| Comando | Resultado esperado |
| --- | --- |
| `curl -i http://localhost:8080/inexistente` | `404 Not Found` |
| `curl -i -X POST http://localhost:8080/health` | `405 Method Not Allowed`, com `Allow: GET, HEAD` |
| `curl -I http://localhost:8080/health` | `200 OK`, header JSON e nenhum corpo |

O roteador padrão associa `HEAD` à rota `GET`. As respostas de erro do roteador são texto simples nesta etapa.

Para verificar porta ocupada, mantenha o primeiro servidor rodando e execute `go run .` em outro terminal. A segunda instância deve encerrar com um erro contendo `bind: address already in use`. O primeiro servidor continua funcionando.

Verificações de código:

```sh
gofmt -l main.go
go vet ./...
go build -o bin/nimbusbank .
```

`gofmt -l` não deve listar arquivos; `go vet` procura problemas comuns; `go build` gera um executável em `bin/`, ignorado pelo Git. A validação HTTP é manual neste passo. Ainda não há testes automatizados; `go test ./...` informará isso.

## Segundo passo: listagem de clientes

`GET /customers` responde `200 OK` e um array JSON com um cliente fictício:

```sh
curl -i http://localhost:8080/customers
```

```json
[{"id":1,"name":"Ana Silva","email":"ana@example.com"}]
```

Se o servidor já estava rodando antes da alteração, encerre com `Ctrl+C` e execute `go run .` novamente. Esse comando não recarrega o código automaticamente.

Leia as novas partes de `main.go`:

1. `type Customer struct` define um tipo nomeado. Cada cliente tem `ID` inteiro e `Name` e `Email` do tipo string. As tags definem os nomes dos campos no JSON.
2. `[]Customer` é um slice: uma sequência de valores do tipo `Customer`. O literal dentro das chaves cria nossa lista com um elemento.
3. `listCustomersHandler` cria os dados fictícios a cada requisição. Ainda não há cadastro nem armazenamento compartilhado.
4. O encoder transforma o slice em um array JSON (`[...]`) e cada struct em um objeto (`{...}`).
5. `mux.HandleFunc("GET /customers", listCustomersHandler)` conecta a rota ao handler.

Dados fixos ajudam a estudar o contrato da API. Uma coleção mutável permitiria cadastrar clientes, mas exigiria proteger leituras e escritas concorrentes; isso será tratado no passo de cadastro. Persistência em PostgreSQL virá na Fase 3.

Para experimentar uma coleção vazia, substitua temporariamente o literal por `customers := []Customer{}`, reinicie o servidor e consulte a rota: a resposta será `[]`. Já `var customers []Customer` declara um slice nil, que o encoder representa como `null`. Depois restaure o exemplo com o cliente. Essa distinção importa para consumidores que esperam sempre um array.

Verifique também:

```sh
curl -i -X POST http://localhost:8080/customers
curl -I http://localhost:8080/customers
curl -i http://localhost:8080/health
```

Resultados esperados: `405` (cadastro ainda não implementado), `200` sem corpo e o health check original com `200` e `{"status":"ok"}`.

Em uma entrevista, explique como uma struct representa um registro, como um slice representa uma coleção e por que os nomes exportados em Go podem ser diferentes dos nomes no JSON. Em uma API profissional, o mesmo contrato de listagem pode ser alimentado por consultas ao banco.

## Uso profissional e entrevista

Health checks permitem que ferramentas verifiquem se um serviço responde. Este endpoint demonstra a disponibilidade HTTP do processo, sem verificar banco de dados ou outras dependências.

Para praticar, explique com suas palavras:

- Qual é a diferença entre servidor, roteador e handler?
- Por que o header é definido antes de escrever o corpo?
- Como método e caminho determinam a resposta `200`, `404` ou `405`?
- Por que responder `200` em `/health` não garante que todas as operações do sistema funcionem?

Antes do próximo passo, execute os comandos, acompanhe o fluxo em `main.go` e valide seu entendimento. Não use senhas ou tokens no código; `.gitignore` ajuda a evitar inclusão acidental de arquivos, mas não detecta secrets escritos em outros arquivos.
