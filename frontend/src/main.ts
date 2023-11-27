import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import vSelect from "vue-select";
import "vue-select/dist/vue-select.css";

import App from './App.vue'
import router from './router'
import axios from "axios";

axios.defaults.headers.common['Access-Control-Allow-Origin'] = '*'

const app = createApp(App)

app.component("v-select", vSelect);
app.use(createPinia())
app.use(router)

app.mount('#app')
