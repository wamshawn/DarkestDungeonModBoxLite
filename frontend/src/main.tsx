import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import {DevSupport} from "@react-buddy/ide-toolbox";
import {ComponentPreviews, useInitial} from "./dev";

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <DevSupport ComponentPreviews={ComponentPreviews}
                    useInitialHook={useInitial}
        >
            <App/>
        </DevSupport>
    </React.StrictMode>
)
