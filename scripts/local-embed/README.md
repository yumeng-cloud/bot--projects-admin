# 本地 Embedding Sidecar（CPU）

OpenAI 兼容 `POST /v1/embeddings`，默认模型 `all-MiniLM-L6-v2`（**384** 维），监听 **8081**。  
与 Go `embedder` 默认 `EMBED_BASE_URL=http://127.0.0.1:8081/v1` 对齐；生产只改 endpoint/model/dims/key。

## 启动

```bash
cd scripts/local-embed
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
python server.py
# 或: uvicorn server:app --host 127.0.0.1 --port 8081
```

首次会下载模型（约几十 MB）。若 pip/torch 失败，换 CPU wheel 或镜像后再装。

## Smoke

```bash
curl -s http://127.0.0.1:8081/v1/embeddings \
  -H 'Content-Type: application/json' \
  -d '{"model":"all-MiniLM-L6-v2","input":["hello semantic search"]}' | head -c 200
```

健康检查：`curl -s http://127.0.0.1:8081/health`
