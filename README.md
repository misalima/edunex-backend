# EduNex

EduNex é uma aplicação web para coordenadores pedagógicos, que centraliza e facilita as tarefas do dia a dia, como acompanhamento de desempenho dos estudantes, gerenciamento de planos de aula e integração com inteligência artificial para apoio na tomada de decisão.

## Funcionalidades Principais

- Autenticação de usuários (coordenadores, professores, etc.)
- Upload e gerenciamento de planos de aula com armazenamento em nuvem (Supabase)
- Listagem e download seguro de planos via URLs assinadas
- Integração futura com IA para análise de planos e desempenho acadêmico
- Agenda integrada e acompanhamento de atividades (em desenvolvimento)

## Tecnologias Utilizadas

- Backend: Go (Golang) com arquitetura hexagonal
- Banco de dados: PostgreSQL
- Armazenamento de arquivos: Supabase Storage
- Frontend: Next.js (em desenvolvimento)
- Autenticação: JWT
- Container de dependências com inicialização lazy e thread-safe

## Como Rodar o Projeto

2. Inicie o banco de dados PostgreSQL e execute o script `init.sql` para criar as tabelas.

3. Compile e rode o backend:

```bash
go run cmd/app/main.go
```

4. Use o Insomnia/Postman para testar os endpoints (ex: upload de planos, autenticação).

## Estrutura do Projeto

- `internal/core`: Domínio e regras de negócio
- `internal/api`: Handlers, container e rotas
- `internal/infra`: Infraestrutura (banco, storage, segurança)
- `cmd/app`: Ponto de entrada da aplicação

## Próximos Passos

- Desenvolvimento do frontend em Next.js
- Implementação da análise de planos com IA
- Expansão das funcionalidades de agenda e acompanhamento

## Contato

Para dúvidas ou contribuições, entre em contato: seu-email@exemplo.com

---

EduNex © 2026 - Projeto pessoal de coordenação pedagógica
```