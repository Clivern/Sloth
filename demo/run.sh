#!/usr/bin/env bash
# Run all demo agents at once. Ctrl+C stops them and deregisters their instances.

cd "$(dirname "$0")"

export SLOTH_URL=http://localhost:8000
export SLOTH_WORKSPACE="<workspace-id>"
export SLOTH_ACCESS_KEY="<access-key>"
export SLOTH_AGENT_SECRET="<assistant-identity-secret>"
export OPENROUTER_API_KEY="<openrouter-key>"
export ASSISTANT_MODEL=deepseek/deepseek-v4-flash
export PYTHONUNBUFFERED=1

# Keep real values in run.local.sh, which git ignores.
[ -f run.local.sh ] && source run.local.sh

agents=(assistant:9200 weather:9201 stock:9202 users:9203)

for agent in "${agents[@]}"; do
  name=${agent%%:*}
  port=${agent##*:}
  uv run uvicorn "$name:app" --port "$port" > >(sed "s/^/[$name] /") 2>&1 &
done

trap 'kill $(jobs -p) 2>/dev/null' INT TERM
wait
