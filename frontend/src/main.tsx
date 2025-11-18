import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import {DevSupport} from "@react-buddy/ide-toolbox";
import {ComponentPreviews, useInitial} from "./dev";
import {App, ConfigProvider, theme} from 'antd';
import {RouterProvider} from "react-router-dom";
import router from "./router/router";


const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    // <React.StrictMode>
        <DevSupport ComponentPreviews={ComponentPreviews}
                    useInitialHook={useInitial}
        >
            <ConfigProvider theme={{algorithm: theme.darkAlgorithm}}>
                <App>
                    <RouterProvider router={router} />
                </App>
            </ConfigProvider>
        </DevSupport>
    // </React.StrictMode>
)
