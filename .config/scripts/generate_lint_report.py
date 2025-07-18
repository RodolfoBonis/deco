import os
import openai
from github import Github
from github import Auth

OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN")
REPO_NAME = os.getenv("REPO_NAME")
PR_NUMBER = os.getenv("PR_NUMBER")

openai.api_key = OPENAI_API_KEY

with open('lint_output.txt', 'r') as file:
    lint_output = file.read().strip()

# Se não houver saída de lint, não comenta nada
if not lint_output:
    print("Nenhum problema de lint encontrado. Nenhum comentário será criado.")
    exit(0)

prompt = f"""
Você é um engenheiro de software sênior revisando um pull request. O CI executou golangci-lint e identificou os seguintes problemas de qualidade de código Go. 

Para cada problema, gere um comentário técnico claro e objetivo, explicando:

* **🔍 Descrição:** O que está errado e por que é importante corrigir
* **📍 Localização:** Arquivo e linha onde o problema ocorre
* **🛠️ Solução:** Como corrigir, incluindo exemplos de código quando relevante
* **⚡ Prioridade:** Alta/Média/Baixa baseada no impacto

**Problemas encontrados pelo golangci-lint:**
```
{lint_output}
```

Formate sua resposta como uma lista numerada em markdown, agrupando problemas similares quando possível. Seja conciso mas informativo.
"""

response = openai.chat.completions.create(
    model="gpt-4o-mini",
    messages=[
        {"role": "system", "content": prompt},
    ],
)

detailed_report = response.choices[0].message.content.strip()

auth = Auth.Token(GITHUB_TOKEN)
git = Github(auth=auth)
repo = git.get_repo(REPO_NAME)
pull_request = repo.get_pull(int(PR_NUMBER))

comment_body = f"### 🔍 Problemas de Lint encontrados pelo CI\n\n{detailed_report}\n\n**💡 Sugestões:**\n\n- Corrija os problemas apontados para garantir a qualidade e padronização do código.\n- Execute `make lint` localmente para validar antes de subir novas alterações.\n- Use `make lint-fix` para corrigir automaticamente alguns problemas de formatação."

pull_request.create_issue_comment(comment_body) 