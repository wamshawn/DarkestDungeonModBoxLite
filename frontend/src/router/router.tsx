import {createHashRouter} from "react-router-dom";
import Plans from "../Application/Plans";
import Modules from "../Application/Modules";
import Settings from "../Application/Settings";
import Application from "../Application";
import ErrorBoundary from "../Application/ErrorBoundary";
import Steam from "../Application/Steam";


const router = createHashRouter([
    {
        path: "/",
        element: <Application/>,
        errorElement: <ErrorBoundary/>,
        children: [
            {
                path: "plans",
                element: <Plans/>
            },
            {
                path: "modules",
                element: <Modules/>
            },
            {
                path: "steam",
                element: <Steam/>
            },
            {
                path: "settings",
                element: <Settings/>,
            }
        ]
    },
]);

export default router