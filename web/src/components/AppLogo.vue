<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 512 512"
    fill="currentColor"
    :role="alt ? 'img' : 'presentation'"
    :aria-label="alt || undefined"
    :aria-hidden="alt ? undefined : true"
  >
    <path fill-rule="evenodd" d="M16 256A240 216 0 1 0 496 256A240 216 0 1 0 16 256ZM52 286A204 156 0 1 0 460 286A204 156 0 1 0 52 286ZM97.1 312.6A78 42 -28 1 0 234.9 239.4A78 42 -28 1 0 97.1 312.6ZM277.1 239.4A78 42 28 1 0 414.9 312.6A78 42 28 1 0 277.1 239.4ZM167 264A17 17 0 1 0 201 264A17 17 0 1 0 167 264ZM311 264A17 17 0 1 0 345 264A17 17 0 1 0 311 264ZM222 336A34 22 0 1 0 290 336A34 22 0 1 0 222 336Z" />
    <path d="M214 382Q256 410 298 382" fill="none" stroke="currentColor" stroke-width="16" stroke-linecap="round" />
  </svg>
</template>

<script setup>
defineProps({
  alt: {
    type: String,
    default: ''
  }
})
</script>
