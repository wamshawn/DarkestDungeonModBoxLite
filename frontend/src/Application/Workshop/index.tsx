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

type WorkshopModule = {
    module: box.WorkshopModule;
    syncing: boolean;
    process: number;
}

type WorkshopModuleSync = {
    total: number;
    current: number;
    processing: boolean;
}

type State = {
    loading: boolean,
    modules: Array<WorkshopModule>,
    syncing: WorkshopModuleSync,
}

const state = proxy<State>({
    loading: false,
    modules: [],
    syncing: {
        total:0,
        current:0,
        processing: false,
    },
});

const Index = () => {
    const {notification} = App.useApp();
    const snap = useSnapshot(state)
    useAsyncEffect(async () => {
       await reflush()
    }, [])

    const reflush = async () => {
        if (state.syncing.processing) {
            notification.warning({message:"警告", description: "请等待同步结束", placement: "bottomRight", duration: 3000})
            return;
        }
        state.modules = new Array<WorkshopModule>();
        state.syncing.current = 0;
        state.syncing.total = 0;
        state.syncing.processing = false;
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
        if (r.value().length > 0 ) {
            state.modules = r.value().map((item) => {
                return {
                    module: item,
                    syncing: false,
                    process: 0,
                }
            })
        }
    }



    const syncing = async () => {
        if (state.modules.length == 0) {
            notification.warning({message:"同步", description: "无新模组待同步", placement: "bottomRight", duration: 3000})
            return;
        }
        state.modules.forEach((item) => {
            if (!item.module.synced) {
                state.syncing.total += 1;
                item.syncing = true;
            }
        })
        if (state.syncing.total === 0) {
            notification.warning({message:"同步", description: "无新模组待同步", placement: "bottomRight", duration: 3000})
            return;
        }
        state.syncing.processing = true;
        state.modules.forEach(item => {
            const module = item.module;
            if (item.module.synced) {
                return
            }
            item.syncing = true;
            state.syncing.current += 1;
            // const r = await rpc<void, Array<box.WorkshopModule>>(ListWorkshopModules)
        })

        state.syncing.processing = false;
    }

    const cancel = async () => {
        // call cancel


        state.modules.forEach(item => {
            item.syncing = false;
        })

        state.syncing.processing = false;
        state.syncing.current = 0;
        state.syncing.total = 0;
    }

    return (
        <Flex>
            <Spin spinning={snap.loading} size={"large"} fullscreen={true}/>
            <Flex wrap gap="middle" >
                {snap.modules.map((item, i) => {
                    const module = item.module
                    let tags: readonly string[] = []
                    if (module.tags && module.tags.length > 0) {
                        tags = module.tags
                    }
                    return (
                        <Spin key={`spin_${module.id}`} spinning={item.syncing} size={"small"}>
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