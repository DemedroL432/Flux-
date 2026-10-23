import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)

/* Global ripple on button press (Apple micro-interaction) — delegated, no per-component code */
document.addEventListener('pointerdown', (e) => {
    const target = (e.target as HTMLElement)?.closest('.btn') as HTMLElement | null
    if (!target || target.hasAttribute('disabled')) return
    const rect = target.getBoundingClientRect()
    const size = Math.max(rect.width, rect.height)
    const span = document.createElement('span')
    span.className = 'ripple'
    span.style.width = span.style.height = `${size}px`
    span.style.left = `${e.clientX - rect.left - size / 2}px`
    span.style.top = `${e.clientY - rect.top - size / 2}px`
    target.appendChild(span)
    span.addEventListener('animationend', () => span.remove())
})
