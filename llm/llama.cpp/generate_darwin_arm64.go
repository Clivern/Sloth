// Copyright 2023 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package llm

//go:generate git submodule init

//go:generate git submodule update --force ggml
//go:generate cmake -S ggml -B ggml/build/metal -DLLAMA_METAL=on -DLLAMA_ACCELERATE=on -DLLAMA_K_QUANTS=on -DCMAKE_SYSTEM_PROCESSOR=arm64 -DCMAKE_OSX_ARCHITECTURES=arm64 -DCMAKE_OSX_DEPLOYMENT_TARGET=11.0
//go:generate cmake --build ggml/build/metal --target server --config Release

//go:generate git submodule update --force gguf
//go:generate cmake -S gguf -B gguf/build/metal -DLLAMA_METAL=on -DLLAMA_ACCELERATE=on -DLLAMA_K_QUANTS=on -DCMAKE_SYSTEM_PROCESSOR=arm64 -DCMAKE_OSX_ARCHITECTURES=arm64 -DCMAKE_OSX_DEPLOYMENT_TARGET=11.0
//go:generate cmake --build gguf/build/metal --target server --config Release
