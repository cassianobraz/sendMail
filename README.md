# sendMail

Aplicacao Go simples para envio de e-mail HTML via SMTP, usando template e imagem embarcados no binario com `embed.FS`.

O projeto monta uma mensagem em HTML, incorpora a imagem `gopher.png` por Content-ID e envia o e-mail utilizando o pacote `gomail`.

## Sumario

- [Visao geral](#visao-geral)
- [Stack](#stack)
- [Arquitetura](#arquitetura)
- [Estrutura de pastas](#estrutura-de-pastas)
- [Variaveis de ambiente](#variaveis-de-ambiente)
- [Como executar](#como-executar)
- [Como funciona o envio](#como-funciona-o-envio)
- [Template do e-mail](#template-do-e-mail)
- [Hot reload com Air](#hot-reload-com-air)
- [Comandos uteis](#comandos-uteis)
- [Boas praticas e pontos de atencao](#boas-praticas-e-pontos-de-atencao)
- [Melhorias futuras](#melhorias-futuras)

## Visao geral

Este repositorio contem um executavel Go localizado em `cmd/api`. Apesar do nome `api`, a aplicacao atual nao sobe um servidor HTTP. Ela executa um fluxo direto:

1. Carrega as variaveis de ambiente a partir do arquivo `.env`.
2. Cria um dialer SMTP com host, porta, usuario e senha.
3. Renderiza o template HTML embarcado.
4. Anexa a imagem embarcada como recurso inline.
5. Envia a mensagem para o proprio e-mail configurado.

O comportamento atual e adequado para estudo, prova de conceito ou base inicial para um servico de disparo de e-mails.

## Stack

- Go `1.26.4`
- `github.com/joho/godotenv` para carregar variaveis de ambiente
- `gopkg.in/gomail.v2` para montagem e envio de e-mails via SMTP
- `embed` da biblioteca padrao para empacotar assets no binario
- Air para hot reload em ambiente de desenvolvimento

## Arquitetura

```text
cmd/
  api/
    main.go                  # Ponto de entrada da aplicacao
  initializers/
    loadEnvVariables.go      # Carregamento do arquivo .env

internal/
  template/
    embed.go                 # Registro dos arquivos embarcados
    mail.html                # Template HTML do e-mail
    gopher.png               # Imagem inline usada no e-mail
```

### Responsabilidades principais

| Arquivo                                | Responsabilidade                                                                      |
| -------------------------------------- | ------------------------------------------------------------------------------------- |
| `cmd/api/main.go`                      | Orquestra o envio do e-mail, configura o SMTP, renderiza o HTML e dispara a mensagem. |
| `cmd/initializers/loadEnvVariables.go` | Carrega variaveis do arquivo `.env` usando `godotenv`.                                |
| `internal/template/embed.go`           | Declara os arquivos embarcados com `//go:embed`.                                      |
| `internal/template/mail.html`          | Define o corpo HTML enviado por e-mail.                                               |

## Estrutura de pastas

```text
.
├── cmd
│   ├── api
│   │   └── main.go
│   └── initializers
│       └── loadEnvVariables.go
├── internal
│   └── template
│       ├── embed.go
│       ├── gopher.png
│       └── mail.html
├── tmp
│   └── build-errors.log
├── .air.toml
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Variaveis de ambiente

Crie um arquivo `.env` na raiz do projeto com base no arquivo `.env.example`:

```env
HOST="smtp.gmail.com"
USER_EMAIL="seu-email@gmail.com"
PASSWORD_EMAIL="sua-senha-ou-app-password"
```

### Descricao

| Variavel         | Obrigatoria | Exemplo               | Descricao                                                                    |
| ---------------- | ----------- | --------------------- | ---------------------------------------------------------------------------- |
| `HOST`           | Sim         | `smtp.gmail.com`      | Servidor SMTP usado no envio.                                                |
| `USER_EMAIL`     | Sim         | `usuario@gmail.com`   | Conta usada como remetente. No fluxo atual tambem e usada como destinatario. |
| `PASSWORD_EMAIL` | Sim         | `abcd efgh ijkl mnop` | Senha SMTP ou app password da conta.                                         |

Para Gmail, o recomendado e usar uma senha de app, nao a senha principal da conta.

## Como executar

### 1. Clone o repositorio

```bash
git clone https://github.com/cassianobraz/sendMail.git
cd sendMail
```

### 2. Configure o ambiente

```bash
cp .env.example .env
```

Edite o `.env` com as credenciais SMTP corretas.

### 3. Baixe as dependencias

```bash
go mod download
```

### 4. Execute a aplicacao

```bash
go run ./cmd/api
```

Se o envio for concluido, a aplicacao exibira:

```text
Sending message.
```

## Como funciona o envio

O fluxo principal esta em `cmd/api/main.go`.

```go
host := os.Getenv("HOST")
userMail := os.Getenv("USER_EMAIL")
password := os.Getenv("PASSWORD_EMAIL")

dialer := gomail.NewDialer(host, 587, userMail, password)
```

A aplicacao usa a porta `587`, comum para SMTP com STARTTLS. Em seguida, a mensagem e criada:

```go
msg := gomail.NewMessage()
msg.SetHeader("From", userMail)
msg.SetHeader("To", userMail)
msg.SetHeader("Subject", "Sending with mail Go")
msg.SetBody("text/html", getBody())
```

Hoje o remetente e o destinatario sao o mesmo e-mail configurado em `USER_EMAIL`. Para enviar para outro endereco, altere o header `To`.

A imagem `gopher.png` e lida do filesystem embarcado e vinculada ao HTML por Content-ID:

```go
msg.Embed("gopher.png", gomail.SetCopyFunc(func(w io.Writer) error {
    _, err := w.Write(img)
    return err
}))
```

No template HTML, ela e referenciada assim:

```html
<img src="cid:gopher.png" alt="gopher" />
```

## Template do e-mail

O corpo do e-mail fica em:

```text
internal/template/mail.html
```

Ele e carregado por `text/template`:

```go
t := htmlTemplate.Must(
    htmlTemplate.ParseFS(templateFS.Files, "mail.html"),
)
```

Atualmente o template nao recebe dados dinamicos. Caso seja necessario personalizar nome, assunto, conteudo ou links, o proximo passo natural e criar uma struct de dados e passa-la para `Execute`.

Exemplo de evolucao:

```go
type MailData struct {
    Name string
    Link string
}

t.Execute(&buff, MailData{
    Name: "Cassiano",
    Link: "https://example.com",
})
```

## Hot reload com Air

O projeto possui configuracao para Air em `.air.toml`.

Instale o Air:

```bash
go install github.com/air-verse/air@latest
```

Execute:

```bash
air
```

A configuracao atual compila o binario em:

```text
./bin/api.exe
```

E monitora alteracoes em arquivos com extensoes:

```text
go, tpl, tmpl, html, env
```

## Comandos uteis

### Executar

```bash
go run ./cmd/api
```

### Rodar validacao dos pacotes

```bash
go test ./...
```

Estado atual:

```text
github.com/cassianobraz/sendMail/cmd/api          [no test files]
github.com/cassianobraz/sendMail/cmd/initializers [no test files]
github.com/cassianobraz/sendMail/internal/template [no test files]
```

### Compilar

```bash
go build -o ./bin/api.exe ./cmd/api
```

### Organizar dependencias

```bash
go mod tidy
```

## Boas praticas e pontos de atencao

### Credenciais

Nunca versionar o arquivo `.env`. O projeto ja ignora esse arquivo no `.gitignore`.

Use senhas de app ou credenciais especificas para SMTP. Evite usar a senha principal da conta de e-mail.

### Tratamento de erros

O fluxo atual usa `panic` em falhas de leitura de asset ou envio de e-mail. Para um servico real, o ideal e retornar erros tratados, registrar logs estruturados e evitar encerrar o processo de forma abrupta.

### Configuracao SMTP

A porta `587` esta fixa no codigo. Para aumentar flexibilidade, a porta poderia virar uma variavel de ambiente:

```env
SMTP_PORT="587"
```

### Destinatario

Atualmente o e-mail e enviado de `USER_EMAIL` para `USER_EMAIL`. Em um uso real, o destinatario deveria vir de configuracao, parametro, fila, banco de dados ou payload HTTP.

### Template dinamico

O template HTML e renderizado sem dados dinamicos. A estrutura ja permite evoluir para dados personalizados com `text/template`.

## Melhorias futuras

- Criar camada `internal/mailer` para isolar a regra de envio.
- Criar struct de configuracao para SMTP.
- Validar variaveis obrigatorias no startup.
- Permitir destinatario, assunto e dados do template por parametro.
- Adicionar suporte a templates dinamicos.
- Adicionar testes unitarios.
- Substituir `panic` por tratamento de erro mais controlado.
- Parametrizar porta SMTP.
- Adicionar logs com contexto.
- Transformar o executavel em API HTTP, caso o objetivo seja expor um endpoint de envio.
