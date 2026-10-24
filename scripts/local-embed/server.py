#!/usr/bin/env python3
"""Minimal OpenAI-compatible /v1/embeddings for local CPU (sentence-transformers).

Default model: sentence-transformers/all-MiniLM-L6-v2 (384-d).
Listen: 127.0.0.1:8081  → matches Go embedder DefaultBaseURL.
"""
from __future__ import annotations

import os
from typing import List, Union

import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

MODEL_ID = os.environ.get("EMBED_MODEL", "all-MiniLM-L6-v2")
HOST = os.environ.get("EMBED_HOST", "127.0.0.1")
PORT = int(os.environ.get("EMBED_PORT", "8081"))

app = FastAPI(title="local-embed", version="0.1.0")
_model = None


def get_model():
    global _model
    if _model is None:
        from sentence_transformers import SentenceTransformer

        name = MODEL_ID
        if "/" not in name:
            name = f"sentence-transformers/{name}"
        _model = SentenceTransformer(name)
    return _model


class EmbedRequest(BaseModel):
    model: str = Field(default=MODEL_ID)
    input: Union[str, List[str]]
    dimensions: int | None = None


@app.get("/health")
def health():
    return {"ok": True, "model": MODEL_ID}


@app.post("/v1/embeddings")
@app.post("/embeddings")
def embeddings(req: EmbedRequest):
    texts = [req.input] if isinstance(req.input, str) else list(req.input)
    if not texts:
        return {"object": "list", "data": [], "model": req.model or MODEL_ID, "usage": {"prompt_tokens": 0, "total_tokens": 0}}
    try:
        model = get_model()
        vecs = model.encode(texts, normalize_embeddings=True)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)) from e
    data = []
    for i, v in enumerate(vecs):
        emb = v.tolist()
        if req.dimensions and req.dimensions > 0 and req.dimensions < len(emb):
            emb = emb[: req.dimensions]
        data.append({"object": "embedding", "embedding": emb, "index": i})
    return {
        "object": "list",
        "data": data,
        "model": req.model or MODEL_ID,
        "usage": {"prompt_tokens": sum(len(t) for t in texts), "total_tokens": sum(len(t) for t in texts)},
    }


if __name__ == "__main__":
    print(f"local-embed model={MODEL_ID} http://{HOST}:{PORT}/v1/embeddings", flush=True)
    uvicorn.run(app, host=HOST, port=PORT, log_level="info")
