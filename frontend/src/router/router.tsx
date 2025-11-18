import {createHashRouter} from "react-router-dom";
import Plans from "../Application/Plans";
import Modules from "../Application/Modules";
import Settings from "../Application/Settings";
import Application from "../Application";


const router = createHashRouter([
    {
        path: "/",
        element: <Application/>,
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
                path: "settings",
                element: <Settings/>,
            }
        ]
    },
]);

export default router