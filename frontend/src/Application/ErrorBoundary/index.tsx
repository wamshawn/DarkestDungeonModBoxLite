import React from "react";
import {isRouteErrorResponse, useNavigate, useRouteError} from "react-router-dom";
import {Button, Result, Typography} from "antd";
import {CloseCircleOutlined} from "@ant-design/icons";

const { Paragraph, Text } = Typography;

type RouteErrorCause = {
    message: string
    stack: string
}

type RouteError = {
    status: number
    statusText: string
    data: string
    error: RouteErrorCause
}

const Index:React.FC<{}> = () => {
    const re = useRouteError();
    const navigate = useNavigate();
    if (isRouteErrorResponse(re)) {
        const error = re as RouteError
        return (
            <Result
                status="error"
                title="系统错误"
                // @ts-ignore
                subTitle={error.statusText}
                extra={
                    <Button type="primary" key="console" onClick={()=>{navigate("/plans")}}>
                        返回
                    </Button>
                }
            >
                <div className="desc">
                    <Paragraph>
                        <Text
                            strong
                            style={{
                                fontSize: 16,
                            }}
                        >
                            具体遇到的问题:
                        </Text>
                    </Paragraph>
                    <Paragraph>
                        <CloseCircleOutlined style={{color:"red"}} /> {error.data}
                    </Paragraph>
                    <Paragraph>
                        <CloseCircleOutlined style={{color:"red"}} /> {error.error.message}
                    </Paragraph>
                    <Paragraph>
                        <CloseCircleOutlined style={{color:"red"}} /> {error.error.stack}
                    </Paragraph>
                </div>
            </Result>
        )
    }
    return (
        <Result
            status="error"
            title="系统错误"
            extra={
                <Button type="primary" key="console" onClick={()=>{navigate("/plans")}}>
                    返回
                </Button>
            }
        >
            <div className="desc">
                <Paragraph>
                    <Text
                        strong
                        style={{
                            fontSize: 16,
                        }}
                    >
                        具体遇到的问题:
                    </Text>
                </Paragraph>
                <Paragraph>
                    <CloseCircleOutlined style={{color:"red"}} /> {"" + re}
                </Paragraph>
            </div>
        </Result>
    )
}

export default Index;