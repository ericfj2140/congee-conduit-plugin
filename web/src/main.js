import { mount } from 'svelte';
import './app.css';
import App from './App.svelte';

document.documentElement.classList.add('dark');

mount(App, { target: document.getElementById('app') });
