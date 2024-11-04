<script setup>
import { ref, onMounted } from 'vue'

const notes = ref([])

const fetchData = async () => {
  const response = await fetch('/api/notes')
  notes.value = await response.json()
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <header>
    <img alt="Vue logo" class="logo" src="./assets/logo.svg" width="125" height="125" />
    <div class="wrapper">
      <div v-for="note in notes" :key="note.id">
        <h2>{{ note.id}}</h2>
        <p>{{ note.text}}</p>
      </div>  
    </div>
  </header>

  <main>
  </main>
</template>

<style scoped>
/* ... style remains the same ... */
</style>
