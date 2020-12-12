package llm

//go:generate git submodule init

//go:generate git submodule update --force ggml
//go:generate cmake -S ggml -B ggml/build/cpu -DLLAMA_K_QUANTS=on
//go:generate cmake --build ggml/build/cpu --target server --config Release

//go:generate git submodule update --force gguf
//go:generate cmake -S gguf -B gguf/build/cpu -DLLAMA_K_QUANTS=on
//go:generate cmake --build gguf/build/cpu --target server --config Release
