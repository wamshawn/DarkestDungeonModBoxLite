import {App, Badge, Card, Flex, FloatButton, Spin, Tag} from "antd";
import {RetweetOutlined, CopyOutlined} from '@ant-design/icons';
import {proxy} from "valtio/vanilla";
import {box} from "../../../wailsjs/go/models";
import {useAsyncEffect} from "ahooks";
import {useSnapshot} from "valtio/react";
import React from "react";
import {rpc} from "../../../wailsjs/rpc/rpc";
import {ListWorkshopModules} from "../../../wailsjs/go/box/Box";
import {Limage} from "../../components/Limage/Limage";

const {Meta} = Card;


type ModuleSync = {
    id: string;
    sync: boolean;
    process: number;
}

type State = {
    loading: boolean,
    modules: Array<box.WorkshopModule>,
    syncing: Array<ModuleSync>,
}

const state = proxy<State>({
    loading: false,
    modules: [],
    syncing: new Array<ModuleSync>,
});

const Index = () => {
    const {notification} = App.useApp();
    const snap = useSnapshot(state)
    useAsyncEffect(async () => {
       await reflush()
    }, [])

    const reflush = async () => {
        state.modules = new Array<box.WorkshopModule>();
        state.syncing = new Array<ModuleSync>();
        state.loading = true;
        const r = await rpc<void, Array<box.WorkshopModule>>(ListWorkshopModules)
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
        state.modules = r.value()
        state.modules.forEach(module => {
            state.syncing.push({id:module.id, sync: false, process: 0})
        })
    }



    const syncing = async () => {


    }

    return (
        <Flex>
            <Spin spinning={snap.loading} size={"large"} fullscreen={true}/>
            <Flex wrap gap="middle" >
                {snap.modules.map((module, i) => {
                    let tags: readonly string[] = []
                    if (module.tags && module.tags.length > 0) {
                        tags = module.tags
                    }
                    const syncing = snap.syncing.find((v) => v.id === module.id)
                    return (
                        <Spin key={`spin_${module.id}`} spinning={syncing?.sync} size={"small"}>
                            <Badge.Ribbon color={module.synced ? "cyan" : ""}
                                          text={module.synced ? "已同步" : "未同步"}>
                                <Card
                                    key={i}
                                    hoverable
                                    style={{width: 240}}
                                    cover={
                                        <Limage
                                            width={240}
                                            height={240}
                                            alt={module.id}
                                            src={module.icon}
                                        />
                                    }
                                    onClick={async () => {}}
                                >
                                    <Meta title={module.title}
                                          description={<Flex wrap gap="small">{tags.map((tag) => (
                                              <Tag bordered={false} color="cyan">{tag}</Tag>
                                          ))}</Flex>}/>
                                </Card>
                            </Badge.Ribbon>
                        </Spin>
                    )
                })}
            </Flex>
            <FloatButton
                icon={<CopyOutlined />} type="default" style={{ insetBlockEnd: 108 }} tooltip={<div>同步至模组管理器</div>}
                onClick={async () => {await syncing()}}
            />
            <FloatButton
                icon={<RetweetOutlined/>} type="primary" tooltip={<div>刷新</div>}
                onClick={async () => { await reflush() }}
            />
        </Flex>

    );
};

export default Index;