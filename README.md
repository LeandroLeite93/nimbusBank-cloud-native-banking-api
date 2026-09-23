# NimbusBank API

Laboratório educacional de Backend, DevOps e Cloud Computing em Go. Simulará operações bancárias para estudo; não é um banco real nem software pronto para produção.

## Etapa atual

Fase 1, terceiro passo: servidor HTTP local com `GET /health`, `GET /customers` e `POST /customers`. Os clientes ficam em memória, começando com um cliente fictício. Sem dependências externas ou recursos cloud. Contas e operações financeiras serão adicionadas progressivamente, após validação do aprendizado.

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
| `main.go` | Entrada do programa, rotas, handlers e clientes em memória |
| `main_test.go` | Testes de cadastro, validação e concorrência |
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

`gofmt -l` não deve listar arquivos; `go vet` procura problemas comuns; `go build` gera um executável em `bin/`, ignorado pelo Git. Além da validação manual, execute `go test ./...` para os testes automatizados e `go test -race ./...` para detectar acessos concorrentes sem sincronização durante os testes.

## Segundo passo: listagem de clientes

`GET /customers` responde `200 OK` e um array JSON. Logo após iniciar o programa, há um cliente fictício:

```sh
curl -i http://localhost:8080/customers
```

```json
[{"id":1,"name":"Ana Silva","email":"ana@example.com"}]
```

Se o servidor já estava rodando antes da alteração, encerre com `Ctrl+C` e execute `go run .` novamente. Esse comando não recarrega o código automaticamente.

Leia as novas partes de `main.go`:

1. `type Customer struct` define um tipo nomeado. Cada cliente tem `ID` inteiro e `Name` e `Email` do tipo string. As tags definem os nomes dos campos no JSON.
2. `[]Customer` é um slice: uma sequência de valores do tipo `Customer`. O literal dentro das chaves inicializa nossa lista compartilhada com um elemento.
3. `listCustomersHandler` copia a lista sob proteção de um mutex antes de produzir a resposta. Assim, também retorna os clientes cadastrados durante esta execução.
4. O encoder transforma o slice em um array JSON (`[...]`) e cada struct em um objeto (`{...}`).
5. `mux.HandleFunc("GET /customers", listCustomersHandler)` conecta a rota ao handler.

No segundo passo usamos dados fixos para estudar o contrato da API. Agora a coleção é mutável e o mutex protege leituras e escritas concorrentes. Persistência em PostgreSQL virá na Fase 3.

Para experimentar uma coleção vazia, substitua temporariamente a inicialização da variável global por `customers = []Customer{}`, reinicie o servidor e consulte a rota: a resposta será `[]`. Um slice nil seria representado diretamente pelo encoder como `null`; nosso handler usa `make` para que a cópia vazia seja sempre representada como `[]`. Depois restaure o exemplo com o cliente. Essa distinção importa para consumidores que esperam sempre um array.

Verifique também:

```sh
curl -i -X DELETE http://localhost:8080/customers
curl -I http://localhost:8080/customers
curl -i http://localhost:8080/health
```

Resultados esperados: `405` (exclusão ainda não implementada), `200` sem corpo e o health check original com `200` e `{"status":"ok"}`.

Em uma entrevista, explique como uma struct representa um registro, como um slice representa uma coleção e por que os nomes exportados em Go podem ser diferentes dos nomes no JSON. Em uma API profissional, o mesmo contrato de listagem pode ser alimentado por consultas ao banco.

## Terceiro passo: cadastro em memória

Reinicie o servidor após alterar o código (`Ctrl+C`, depois `go run .`). Em outro terminal:

```sh
curl -i -X POST http://localhost:8080/customers \
  -H 'Content-Type: application/json' \
  -d '{"name":"Bruno Lima","email":"bruno@example.com"}'
```

`-X POST` escolhe o método, `-H` declara o formato do corpo e `-d` envia os dados. Esperado no primeiro cadastro após iniciar o servidor:

```text
HTTP/1.1 201 Created
Content-Type: application/json

{"id":2,"name":"Bruno Lima","email":"bruno@example.com"}
```

Consulte `GET /customers` novamente: a lista deve conter Ana e Bruno. Cada POST válido cria um novo cliente e recebe outro ID; não verificamos duplicidade de e-mail neste passo. Ao reiniciar, os cadastros são perdidos e o contador volta a 2.

Como o cadastro funciona:

1. Uma struct de entrada contém somente `Name` e `Email`. O ID pertence à resposta e é gerado pelo servidor.
2. `json.NewDecoder(r.Body).Decode(&input)` lê JSON do corpo e preenche a struct. `&input` fornece o endereço da variável para que o decoder possa modificá-la.
3. `DisallowUnknownFields` rejeita campos desconhecidos, inclusive `id`. Uma segunda leitura deve encontrar `io.EOF` (fim da entrada); isso rejeita dois objetos ou conteúdo extra no mesmo corpo.
4. `strings.TrimSpace` remove espaços nas extremidades. Campos ausentes, vazios ou contendo apenas espaços são rejeitados. Ainda não validamos formato nem unicidade do e-mail.
5. `customersMu.Lock()` protege a geração do ID, o `append` na lista e o incremento do contador como uma única seção crítica. `Unlock()` libera o acesso para outra requisição.
6. `WriteHeader(http.StatusCreated)` envia explicitamente `201` antes do corpo JSON. Erros de entrada usam `http.Error`, com status `400` e mensagem em texto simples.

O GET copia a lista enquanto mantém o mutex e libera a trava antes de escrever pela rede. Isso mantém uma visão consistente dos clientes sem bloquear cadastros enquanto um consumidor recebe sua resposta. As variáveis globais mantêm este exercício pequeno; isolamento do estado e separação de responsabilidades serão abordados na Fase 2.

Teste uma entrada inválida:

```sh
curl -i -X POST http://localhost:8080/customers \
  -H 'Content-Type: application/json' \
  -d '{"name":"   ","email":"bruno@example.com"}'
```

Esperado: `400 Bad Request`. Consulte a lista e confirme que nada foi adicionado. JSON quebrado, corpo vazio e campos desconhecidos também devem retornar `400`.

Execute os testes:

```sh
go test -v ./...
go test -race ./...
```

`main_test.go` usa `httptest` para enviar requisições diretamente aos handlers e inspecionar as respostas, sem ocupar a porta 8080. Verifica cadastro seguido de consulta, limpeza de espaços, entradas inválidas sem alteração de estado e 20 cadastros concorrentes intercalados com consultas. O detector `-race` procura acessos à memória sem sincronização nos caminhos executados; não é uma prova de ausência de todos os problemas de concorrência.

Profissionalmente, validação protege o contrato da API, e sincronização protege dados compartilhados. O mutex coordena apenas este processo: ele não substitui transações no banco nem coordena várias instâncias da API.

Perguntas de entrevista: por que usamos `201` no cadastro e `200` na consulta? O que poderia acontecer se duas requisições lessem o mesmo próximo ID sem trava? Por que dados em memória desaparecem quando o processo termina?

## Uso profissional e entrevista

Health checks permitem que ferramentas verifiquem se um serviço responde. Este endpoint demonstra a disponibilidade HTTP do processo, sem verificar banco de dados ou outras dependências.

Para praticar, explique com suas palavras:

- Qual é a diferença entre servidor, roteador e handler?
- Por que o header é definido antes de escrever o corpo?
- Como método e caminho determinam a resposta `200`, `404` ou `405`?
- Por que responder `200` em `/health` não garante que todas as operações do sistema funcionem?

Antes do próximo passo, execute os comandos, acompanhe o fluxo em `main.go` e valide seu entendimento. Não use senhas ou tokens no código; `.gitignore` ajuda a evitar inclusão acidental de arquivos, mas não detecta secrets escritos em outros arquivos.
