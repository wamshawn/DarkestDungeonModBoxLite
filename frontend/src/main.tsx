import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import 'dayjs/locale/zh-cn';
import {App, ConfigProvider, theme} from 'antd';
import locale from 'antd/locale/zh_CN';
import {RouterProvider} from "react-router-dom";
import router from "./router/router";


const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <ConfigProvider theme={{algorithm: theme.darkAlgorithm}} locale={locale}>
            <App>
                <RouterProvider router={router}/>
            </App>
        </ConfigProvider>
    </React.StrictMode>
)
