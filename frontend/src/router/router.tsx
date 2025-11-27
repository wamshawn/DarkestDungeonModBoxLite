import {createHashRouter} from "react-router-dom";
import Schemas from "../Application/Schemas";
import Modules from "../Application/Modules";
import Settings from "../Application/Settings";
import Application from "../Application";
import ErrorBoundary from "../Application/ErrorBoundary";
import Workshop from "../Application/Workshop";


const router = createHashRouter([
    {
        path: "/",
        element: <Application/>,
        errorElement: <ErrorBoundary/>,
        children: [
            {
                path: "schemas",
                element: <Schemas/>
            },
            {
                path: "modules",
                element: <Modules/>
            },
            {
                path: "workshop",
                element: <Workshop/>
            },
            {
                path: "settings",
                element: <Settings/>,
            }
        ]
    },
]);

export default router