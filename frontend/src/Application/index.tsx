import React, {useState} from "react";
import {
    ReconciliationOutlined,
    AppstoreAddOutlined,
    SettingOutlined,
    GiftOutlined,
} from '@ant-design/icons';

import {Outlet, useLocation, useNavigate} from "react-router-dom";
import {App, Layout, Menu, MenuProps, Spin} from 'antd';
import {useAsyncEffect} from "ahooks";
import {proxy} from "valtio/vanilla";
import {useSnapshot} from "valtio/react";
import {rpc} from "../../wailsjs/rpc/rpc";
import {Open} from "../../wailsjs/go/box/Box";
import {box} from "../../wailsjs/go/models";
import {WindowMaximise} from "../../wailsjs/runtime";
const { Header, Content, Sider } = Layout;


type MenuItem = Required<MenuProps>['items'][number];

function getItem(
    label: React.ReactNode,
    key: React.Key,
    icon?: React.ReactNode,
    children?: MenuItem[],
): MenuItem {
    return {
        key,
        icon,
        children,
        label,
    } as MenuItem;
}

const items: MenuItem[] = [
    getItem('方案', 'schemas', <ReconciliationOutlined />),
    getItem('模组', 'modules', <AppstoreAddOutlined />),
    getItem('工坊', 'workshop', <GiftOutlined />),
    getItem('设置', 'settings', <SettingOutlined/>),
];

type State = {
    loading: boolean,
    menuKey: string,
}

const state = proxy<State>({
    loading: false,
    menuKey: "plans",
});

const Index: React.FC = () => {
    const {notification} = App.useApp();
    const navigate = useNavigate();
    const snap = useSnapshot(state)

    useAsyncEffect(async () => {
        state.loading = true;
        const r = await rpc<void, box.Settings>(Open)
        state.loading = false;
        if (r.failed()) {
            const errs = r.cause()
            for (let i = 0; i < errs.length; i++) {
                notification.error({
                    message: errs[i].error,
                    description: errs[i].description,
                    placement: "bottomRight",
                    role: "alert"
                })
            }
            return;
        }
        // WindowMaximise();
        const settings = r.value()
        if (settings.game == "") {
            state.menuKey = 'settings';
            navigate('settings');
            return;
        }
        state.menuKey = 'schemas';
        navigate('schemas');
        return;
    }, [])

    return (
        <Layout style={{minHeight: '100vh'}}>
            <Spin spinning={snap.loading} size={"large"} fullscreen/>
            <Sider collapsed={true} collapsedWidth={60}
                   style={{
                       overflow: 'auto',
                       height: '100vh',
                       position: 'sticky',
                       insetInlineStart: 0,
                       top: 0,
                       bottom: 0,
                       scrollbarWidth: 'thin',
                   }}
            >
                <Menu
                    theme="dark"
                    selectedKeys={[snap.menuKey]} mode="inline" items={items}
                    onClick={({key}) => {
                        if (key === "schemas") {
                            state.menuKey = 'schemas';
                            navigate("schemas")
                        } else if (key === "modules") {
                            state.menuKey = 'modules';
                            navigate("modules")
                        } else if (key === "workshop") {
                            state.menuKey = 'workshop';
                            navigate("workshop")
                        } else if (key === "settings") {
                            state.menuKey = 'settings';
                            navigate("settings")
                        }
                    }}

                />
            </Sider>
            <Layout>
                <Content style={{ margin: '30px' }}>
                    <Outlet/>
                </Content>
            </Layout>
        </Layout>
    );
};

export default Index;
