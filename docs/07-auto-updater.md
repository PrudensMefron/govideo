# GoVideo — Auto-updater e releases

## Estado implementado

O GoVideo usa o updater nativo do Wails `v3.0.0-beta.20` e consulta
diretamente os releases públicos do próprio remote `origin`:

```text
https://github.com/PrudensMefron/govideo.git
```

O identificador usado pela API é `PrudensMefron/govideo`. O suporte inicial
de release contém dois executáveis:

```text
govideo-linux-amd64
govideo-windows-amd64.exe
```

O updater e o pipeline não oferecem ARM64, macOS, instaladores ou pacotes
Linux nesta etapa.

## Fluxo no aplicativo

O código está isolado em `internal/desktop/update` e continua fora do domínio
e dos jobs de mídia:

```text
Vue / Configurações
        ↓
DesktopService
        ↓
internal/desktop/update
        ↓
Wails updater / provider GitHub
        ↓
PrudensMefron/govideo Releases
```

`main.buildVersion` tem o valor seguro `0.0.0-dev` por padrão. O build de
release injeta a versão da tag com linker flags. Builds de desenvolvimento:

- mostram a versão atual nas configurações;
- mantêm o updater desativado;
- não consultam o GitHub automaticamente.

Builds versionadas:

1. abrem o aplicativo sem bloquear a interface;
2. aguardam 15 segundos;
3. executam somente `Updater.Check` em segundo plano;
4. publicam o snapshot `update:updated` para o frontend;
5. mostram a versão disponível nas configurações;
6. ao clicar em `Baixar atualização`, chamam `CheckAndInstall`;
7. usam a janela padrão do Wails para release notes, progresso e reinício.

O fluxo manual `Verificar atualizações` também usa `CheckAndInstall`. O Wails
baixa o artefato, valida o checksum, prepara a troca e somente reinicia depois
da ação explícita do usuário na janela do updater.

As operações de atualização são serializadas. Uma segunda tentativa enquanto
outra estiver em curso retorna erro, em vez de iniciar dois downloads.

## Segurança e verificação

O provider oficial é configurado assim:

```go
github.New(github.Config{
    Repository:    "PrudensMefron/govideo",
    ChecksumAsset: "SHA256SUMS",
})
```

Na `beta.20`, a falta do asset `SHA256SUMS` ou da linha correspondente não é
um erro obrigatório: o provider pode devolver uma release sem verificação. O
GoVideo envolve o provider com uma política própria que exige:

```text
Verification != nil
DigestAlgo == "sha256"
len(Digest) == 32
```

Sem essas condições, a release é recusada antes do download. O digest é
calculado durante o streaming pelo Wails e comparado antes de preparar a troca.

SHA-256 protege contra arquivo incompleto ou diferente do publicado, mas o
checksum está no mesmo release do executável e não constitui uma assinatura
independente. Assinatura Ed25519, Authenticode e proteção contra comprometimento
do canal de release permanecem como hardening futuro.

Não há GitHub PAT embutido no aplicativo. O repositório é público e usa a API
anônima. Um token compilado no binário não seria secreto.

## Como o Wails aplica a atualização

Depois da verificação, o Wails:

1. mantém o novo executável em staging temporário;
2. reinicia o executável atual em modo helper, sem binário auxiliar separado;
3. pede o encerramento normal da aplicação;
4. aguarda o processo pai sair;
5. preserva um backup;
6. substitui o executável;
7. relança a nova versão;
8. tenta restaurar o backup se a troca ou o relançamento falhar.

No Windows, o helper da `beta.20` usa `HideWindow`, `CREATE_NO_WINDOW`,
`DETACHED_PROCESS` e `CREATE_NEW_PROCESS_GROUP`. A atualização não deve abrir
um terminal. O executável normal também é compilado com `-H windowsgui`.

No Linux, o Wails restaura a permissão executável original. Em ambos os
sistemas, o usuário precisa ter permissão de escrita no diretório que contém o
aplicativo.

## Contrato dos releases

### Tags e versões

- A tag deve seguir `vMAJOR.MINOR.PATCH`, por exemplo `v0.2.0`.
- A versão embutida usa o mesmo valor sem `v`, por exemplo `0.2.0`.
- Tags inválidas fazem o workflow falhar.
- O updater usa comparação SemVer e não instala versão igual ou inferior.
- Releases são criados como draft e precisam de publicação manual.

Não altere os assets de uma release já publicada. Para qualquer correção,
crie uma versão nova.

### Assets obrigatórios

Cada draft recebe exatamente os payloads do updater e seu checksum:

```text
govideo-linux-amd64
govideo-windows-amd64.exe
SHA256SUMS
```

O auto-updater recebe executáveis diretos, não `.zip`, AppImage, NSIS, `.deb`
ou `.rpm`. Esses formatos podem ser acrescentados para primeira instalação em
uma etapa posterior, mas não devem substituir os nomes acima.

## Pipeline CI/CD

O workflow está em `.github/workflows/release.yml`.

### Eventos

- Push em qualquer branch: testes e compilação dos dois sistemas.
- Pull request: testes e compilação dos dois sistemas.
- Execução manual: testes e compilação dos dois sistemas.
- Tag `v*.*.*`: além de testar e compilar, cria ou atualiza um draft release.

### Ferramentas fixadas

```text
Go       1.25.0
Node     24
Frontend npm + package-lock.json
Wails    v3.0.0-beta.20
```

O projeto continua preferindo pnpm no desenvolvimento local. O workflow usa
explicitamente `npm ci` e passa `PACKAGE_MANAGER=npm` ao Wails, conforme o
contrato de CI/CD.

### Etapas

1. valida a tag e deriva a versão;
2. instala dependências com `npm ci`;
3. executa o build TypeScript/Vite com npm;
4. executa `go test -race ./...`;
5. instala o Wails CLI fixado em `beta.20`;
6. compila Linux AMD64 em `ubuntu-24.04`;
7. compila Windows AMD64 em `windows-2025`;
8. injeta a versão usando `-X main.buildVersion=<versão>`;
9. transfere os binários entre jobs como artifacts temporários;
10. confere a presença dos dois payloads;
11. gera e valida `SHA256SUMS`;
12. cria o GitHub Release como draft com release notes automáticas.

Se um draft da mesma tag já existir, o workflow apenas substitui seus assets.
Se a release já estiver publicada, o workflow falha para não alterar um
release imutável. O job de draft recebe `contents: write`; os demais permanecem
com `contents: read`.

## Criando um release

1. confirme que a branch destinada ao release está integrada;
2. crie e envie uma tag:

```sh
git tag v0.2.0
git push origin v0.2.0
```

3. acompanhe o workflow `CI and draft release`;
4. baixe e teste os dois binários do draft;
5. confira `SHA256SUMS`;
6. publique manualmente o draft no GitHub.

O updater consulta apenas releases publicados. Um draft nunca será oferecido
aos usuários, permitindo validar os artefatos antes de promovê-los.

## Restrições de instalação

### Windows

O updater consegue substituir uma instalação gravável pelo usuário. Uma
instalação machine-wide em `Program Files` normalmente exige elevação. Antes de
adotar NSIS como canal principal, deve-se escolher instalação per-user ou
desativar a troca automática para instalações machine-wide.

### Linux

O fluxo funciona para o binário portátil em diretório gravável. O aplicativo
não deve sobrescrever arquivos gerenciados por `.deb`, `.rpm` ou instalados em
`/usr/bin`/`/opt`. AppImage também não é coberto pelo payload binário inicial.

## Testes e próximos hardenings

O pacote possui testes para:

- recusa de release sem checksum;
- recusa de algoritmo diferente de SHA-256;
- aceite de digest SHA-256 com 32 bytes;
- ativação apenas em builds versionadas.

Antes da primeira publicação estável ainda devem ser executados testes reais
`N → N+1` em Windows 11 e Linux, incluindo troca, relançamento, falta de
permissão e checksum inválido.

Hardenings futuros:

- assinatura independente Ed25519 do artefato/manifesto;
- Authenticode para Windows;
- instalador per-user;
- detecção explícita do tipo de instalação;
- canais de prerelease;
- Linux/Windows ARM64;
- packages/instaladores de primeira instalação.

## Referências

- Tutorial oficial do Wails:
  <https://v3.wails.io/tutorials/04-self-update-a-wails-app/>
- Guia de build multiplataforma:
  <https://v3.wails.io/guides/build/cross-platform/>
- Código do updater fixado no projeto:
  <https://github.com/wailsapp/wails/tree/v3.0.0-beta.20/v3/pkg/updater>
- Repositório de demonstração:
  <https://github.com/wailsapp/updater-demo>

Como Wails 3 ainda está em beta, uma atualização do framework deve ser feita
separadamente e acompanhada de nova revisão deste fluxo.
