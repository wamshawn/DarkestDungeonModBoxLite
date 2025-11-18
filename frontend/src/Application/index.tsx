import React, {useState} from "react";
import {
    ReconciliationOutlined,
    AppstoreAddOutlined,
    SettingOutlined,
} from '@ant-design/icons';

import {Outlet, useLocation, useNavigate} from "react-router-dom";
import {App, Avatar, Breadcrumb, Button, Layout, Menu, MenuProps, Spin} from 'antd';
import {useAsyncEffect} from "ahooks";
import {proxy} from "valtio/vanilla";
import {useSnapshot} from "valtio/react";
import {rpc} from "../../wailsjs/rpc/rpc";
import {Open} from "../../wailsjs/go/box/Box";
import {box} from "../../wailsjs/go/models";
const { Header, Content, Footer, Sider } = Layout;


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
    getItem('方案', 'plans', <ReconciliationOutlined />),
    getItem('模组', 'modules', <AppstoreAddOutlined />),
    getItem('设置', 'settings', <SettingOutlined />),
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
    const { notification } = App.useApp();
    const navigate = useNavigate();
    const snap = useSnapshot(state)

    useAsyncEffect(async () => {
        state.loading = true;
        const r = await rpc<void, box.Settings>(Open)
        state.loading = false;
        if (r.failed()) {
            notification.error({message: "错误", description:r.cause(), placement:"bottomRight"})
            return;
        }
        const settings = r.value()
        if (settings.mods == "" || settings.game == "") {
            state.menuKey = 'settings';
            navigate('settings');
            return;
        }
        state.menuKey = 'plans';
        navigate('plans');
        return;
    }, [])

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Spin spinning={snap.loading}  fullscreen />
            <Sider collapsed={true} collapsedWidth={60} >
                <Menu
                    theme="dark"
                    selectedKeys={[snap.menuKey]} mode="inline" items={items}
                    onClick={({key}) => {
                        if (key === "plans") {
                            navigate("plans")
                        } else if (key === "modules") {
                            navigate("modules")
                        } else if (key === "settings") {
                            navigate("settings")
                        }
                    }}

                />
            </Sider>
            <Layout>
                <Content style={{ margin: '16px' }}>
                    <Outlet/>
                </Content>
            </Layout>
        </Layout>
    );
};

export default Index;
