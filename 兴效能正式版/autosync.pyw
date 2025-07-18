from alibabacloud_dingtalk.yida_1_0.client import Client as dingtalkyida_1_0Client
from alibabacloud_tea_openapi import models as open_api_models
from alibabacloud_dingtalk.yida_1_0 import models as dingtalkyida__1__0_models
from alibabacloud_tea_util import models as util_models
from alibabacloud_dingtalk.oauth2_1_0.client import Client as dingtalkoauth2_1_0Client
from alibabacloud_dingtalk.oauth2_1_0 import models as dingtalkoauth_2__1__0_models
from alibabacloud_dingtalk.robot_1_0 import models as dingtalkrobot__1__0_models
from alibabacloud_dingtalk.robot_1_0.client import Client as dingtalkrobot_1_0Client
import os
from traceback import format_exc
from requests import post
from json import load, dumps
from time import time, sleep
from datetime import datetime, timedelta
from functools import wraps

CACHE = os.getenv("LOCALAPPDATA")
LOG = rf"{CACHE if CACHE else '.'}\pkxes.log"

APP = r"APP_D1BC57CD7CRJPZXN07FL"
INFO = r"FORM-6F4AE8B3490040A29A25A576BCC33438OY6S"
WARN = r"FORM-77CB92C91D1E40EABCFEF76C54292AA0T50J"
SUBS = r"FORM-61E5F8B83F1647E3BAB11AA2A1A6844AZUHA"
RUNNING = r"FORM-62C95C6607254C9DB43679F1265697D1SXSC"
RUNSUB = r"FORM-40171AF08DE848738E71BA557B270ED7M5S7"
DUTY = r"FORM-J7966ZA162Q3XXKY5GNVYCDM4C5J3SZGHU98L93"

KEY = r"ding2gb8oi5iswe8mj7q"
SECRET = r"0yu7UqwlRBtHt5XaI24J1JUMGPJMP_X9t6RY3gN4rLjUePna5iWal2P1fk06HwnC"
UID = r"0148582454111130975274"
SYSTOKEN = r"E0D660C1XV9MW6E1DVNCD8EZS8RV22WOH3TXLP"
RENAMER_INFO = {
    "数据日期时间": "dateField_lxt3od6n",
    "综合效率席短信": "textareaField_lxt3od6p",
    "航班正常性与延误详情": "textareaField_lxt3od6r",
    "当前运行概述": "textareaField_lxt5czt2",
    "CTOT推点航班": "textField_lxt3od6v",
    "CTOT推点航班图片": "textField_lxujxpcr",
    "延误1小时以上未起飞航班": "textField_lxt3od6z",
    "延误1小时以上未起飞航班图片": "textField_lxujxpcs",
    "始发": "numberField_lxtsoeu1",
    "放行": "numberField_lxtsoeu2",
    "起飞": "numberField_lxtsoeu3",
    "进港": "numberField_lxtsoeu4",
    "CTOT推点": "numberField_lxtsoeuc",
    "延误未起飞": "numberField_lxtsoeud",
    "大面积航延": "textField_lxt3od71",
    "启动标准": "textField_lxtlrmjl",
    "四地八场": "textField_mb8ok7lr",
    "更新类型": "numberField_lxvgxw12",
    "预警条件": "numberField_lypcbe0c",
    "响应条件一": "numberField_lypcbe0e",
    "响应条件二": "numberField_lypcbe0g",
    "响应条件三": "numberField_lypcbe0i",
}
RENAMER_WARN = {
    "航班号": "textField_ly7cbfcg",
    "日期": "dateField_ly7cbfch",
    "目的地": "textField_lz2cvy1p",
    "机位": "textField_ly7cbfci",
    "登机门": "textField_lz2cvy1o",
    "地服": "textField_lyiiagcu",
    "类型": "textField_ly8a3pkx",
    "描述": "textField_ly7cbfck",
}
RENAMER_RUNNING = {
    "标题": "textField_lz6us3g2",
    "表格": "textField_lz6us3g4",
    "时间": "dateField_lz6us3g8",
    "类型": "numberField_lzcc2qad"
}
RENAMER_RUNSUB = {
    "跑道运行态势": "textField_lz714eu1",
    "待离港航班态势": "textField_lz714eu3",
    "实际/计划 进离港态势": "textareaField_lz7v98ky",
    "机上等待态势": "textField_lz7v98kz",
    "当日执行态势": "textField_lz8gkwcs",
    "截至当前执行态势": "textField_lzea90eq",
    "时间": "dateField_lz714eu5",
}
TABLE = {
    RENAMER_RUNSUB["跑道运行态势"]: {
        "meta": [
            {
                "aliasName": r"方向",
                "dataType": "STRING",
                "alias": "dir",
                "weight": 14
            },
            {
                "aliasName": r"上次起飞",
                "dataType": "STRING",
                "alias": "lt",
                "weight": 21
            },
            {
                "aliasName": r"起飞间隔",
                "dataType": "STRING",
                "alias": "sep",
                "weight": 23
            },
            {
                "aliasName": r"上次落地",
                "dataType": "STRING",
                "alias": "ll",
                "weight": 21
            },
            {
                "aliasName": r"预起/落",
                "dataType": "STRING",
                "alias": "e",
                "weight": 20
            }
        ]
    },
    RENAMER_RUNSUB["待离港航班态势"]: {
        "meta": [
            {
                "aliasName": r"方向",
                "dataType": "STRING",
                "alias": "dir",
                "weight": 13
            },
            {
                "aliasName": r"登机",
                "dataType": "STRING",
                "alias": "asbt",
                "weight": 13
            },
            {
                "aliasName": r"登结",
                "dataType": "STRING",
                "alias": "aebt",
                "weight": 13
            },
            {
                "aliasName": r"关舱",
                "dataType": "STRING",
                "alias": "acct",
                "weight": 13
            },
            {
                "aliasName": r"滑行",
                "dataType": "STRING",
                "alias": "push",
                "weight": 13
            },
            {
                "aliasName": r"滑回",
                "dataType": "STRING",
                "alias": "slibk",
                "weight": 13
            },
            {
                "aliasName": r"待/完冰",
                "dataType": "STRING",
                "alias": "icp",
                "weight": 24
            }
        ]
    },
    RENAMER_RUNSUB["实际/计划 进离港态势"]: {
        "meta": [
            {
                "aliasName": "方向",
                "dataType": "STRING",
                "alias": "dir",
                "weight": 13
            },
            {
                "aliasName": "前1小时",
                "dataType": "STRING",
                "alias": "dep",
                "weight": 20
            },
            {
                "aliasName": "半小时离",
                "dataType": "STRING",
                "alias": "hdep",
                "weight": 22
            },
            {
                "aliasName": "|",
                "dataType": "STRING",
                "alias": "i",
                "weight": 2
            },
            {
                "aliasName": "前1小时",
                "dataType": "STRING",
                "alias": "arr",
                "weight": 20
            },
            {
                "aliasName": "半小时进",
                "dataType": "STRING",
                "alias": "harr",
                "weight": 22
            }
        ]
    },
    RENAMER_RUNSUB["机上等待态势"]: {
        "meta": [
            {
                "aliasName": "方向",
                "dataType": "STRING",
                "alias": "dir",
                "weight": 14
            },
            {
                "aliasName": "未起",
                "dataType": "STRING",
                "alias": "n",
                "weight": 13
            },
            {
                "aliasName": "已起",
                "dataType": "STRING",
                "alias": "a",
                "weight": 13
            },
            {
                "aliasName": "总计",
                "dataType": "STRING",
                "alias": "s",
                "weight": 13
            },
            {
                "aliasName": "平均时长",
                "dataType": "STRING",
                "alias": "mean",
                "weight": 24
            },
            {
                "aliasName": "最长时长",
                "dataType": "STRING",
                "alias": "max",
                "weight": 24
            }
        ]
    },
    RENAMER_RUNSUB["当日执行态势"]: {
        "meta": [
            {
                "aliasName": "架次",
                "dataType": "STRING",
                "alias": "dir",
                "weight": 14
            },
            {
                "aliasName": "计划",
                "dataType": "STRING",
                "alias": "t",
                "weight": 14
            },
            {
                "aliasName": "已执行",
                "dataType": "STRING",
                "alias": "a",
                "weight": 18
            },
            {
                "aliasName": "未执行",
                "dataType": "STRING",
                "alias": "s",
                "weight": 18
            },
            {
                "aliasName": "取消",
                "dataType": "STRING",
                "alias": "c",
                "weight": 12
            },
            {
                "aliasName": "昨日剩余",
                "dataType": "STRING",
                "alias": "n",
                "weight": 21
            }
        ]
    },
    RENAMER_RUNSUB["截至当前执行态势"]: {
        "meta": [
            {
                "aliasName": "架次",
                "dataType": "STRING",
                "alias": "dir",
                "weight": 18
            },
            {
                "aliasName": "西离",
                "dataType": "STRING",
                "alias": "wd",
                "weight": 13
            },
            {
                "aliasName": "东离",
                "dataType": "STRING",
                "alias": "ed",
                "weight": 13
            },
            {
                "aliasName": "离港",
                "dataType": "STRING",
                "alias": "d",
                "weight": 13
            },
            {
                "aliasName": "|",
                "dataType": "STRING",
                "alias": "i",
                "weight": 2
            },
            {
                "aliasName": "西进",
                "dataType": "STRING",
                "alias": "wa",
                "weight": 13
            },
            {
                "aliasName": "东进",
                "dataType": "STRING",
                "alias": "ea",
                "weight": 13
            },
            {
                "aliasName": "进港",
                "dataType": "STRING",
                "alias": "a",
                "weight": 13
            }
        ]
    }
}


TOKEN_TIMESTAMP = 0
EXPIRE = 7200
PATH = r"\\10.154.61.233\运行管理部\兴效能"  # r"."

CONFIG = open_api_models.Config()
CONFIG.protocol = "https"
CONFIG.region_id = "central"


def retry(retries: int = 3, delay: float = 1):
    if retries < 1 or delay <= 0:
        retries = 3
        delay = 1

    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            for i in range(retries + 1):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    if i == retries:
                        update_log(f'"{func.__name__}()" 执行失败，已重试{retries}次')
                        break
                    else:
                        update_log(
                            f"{repr(e)}，{delay}秒后第[{i+1}/{retries}]次重试..."
                        )
                        sleep(delay)

        return wrapper

    return decorator


def update_log(__str: str):
    try:
        with open(LOG, "a", encoding="UTF-8") as output:
            output.write("{} {}\n".format(str(datetime.now())[:19], __str))
    except Exception:
        ...


@retry(5, 0.5)
def clear_log(days: int = 7):
    days = datetime.now() - timedelta(days)
    if os.path.exists(LOG):
        with open(LOG, "r", encoding="UTF-8") as log:
            lines = log.readlines()
        for i, line in enumerate(lines):
            try:
                if datetime.fromisoformat(line[:19]) >= days:
                    break
            except Exception:
                continue
        if i > 0:
            try:
                with open(LOG, "w", encoding="UTF-8") as log:
                    log.writelines(lines[i:])
            except PermissionError:
                ...


def upload_img(path: str):
    # return '@lALPDgCwd_scHXbNArTNBKE'
    url = (
        f"https://oapi.dingtalk.com/media/upload?access_token={get_token()}&type=image"
    )
    retries, returns = 5, ""
    while retries:
        try:
            retries -= 1
            with open(path, "rb") as file:
                files = {"media": file}
                result = post(url, files=files).json()
            if result["errcode"] == 0:
                returns = result["media_id"]
                break
            else:
                update_log(f"图片上传失败: {result['errmsg']}")
        except Exception:
            pass
        sleep(5)
    else:
        update_log("图片上传失败并停止")
    return returns


@retry(100, 2)
def request_token() -> tuple[str, int]:
    client = dingtalkoauth2_1_0Client(CONFIG)

    get_access_token_request = dingtalkoauth_2__1__0_models.GetAccessTokenRequest(
        app_key=KEY, app_secret=SECRET
    )
    try:
        body = client.get_access_token(get_access_token_request).body
        return body.access_token, body.expire_in
    except Exception as err:
        raise Exception(f"TOKEN获取错误（{err.code}）： {err.message}")


def get_token() -> str:
    global TOKEN_TIMESTAMP, EXPIRE, ACCESS_TOKEN
    if time() - TOKEN_TIMESTAMP >= EXPIRE - 60:
        ACCESS_TOKEN, EXPIRE = request_token()
        TOKEN_TIMESTAMP = time()
    return ACCESS_TOKEN


@retry(10, 2)
def upload(json_list: list[str], form: str, name: str = "", app: str = APP, systoken: str = SYSTOKEN) -> None:
    client = dingtalkyida_1_0Client(CONFIG)

    batch_save_form_data_headers = dingtalkyida__1__0_models.BatchSaveFormDataHeaders()
    batch_save_form_data_headers.x_acs_dingtalk_access_token = get_token()
    batch_save_form_data_request = dingtalkyida__1__0_models.BatchSaveFormDataRequest(
        no_execute_expression=False,
        form_uuid=form,
        app_type=app,
        asynchronous_execution=True,
        system_token=systoken,
        keep_running_after_exception=True,
        user_id=UID,
        form_data_json_list=json_list,
    )
    try:
        client.batch_save_form_data_with_options(
            batch_save_form_data_request,
            batch_save_form_data_headers,
            util_models.RuntimeOptions(),
        )
    except Exception as err:
        raise Exception(f"{name}同步错误（{err.code}）： {err.message}")


def update(data: dict, fiid: str, name: str = ""):
    client = dingtalkyida_1_0Client(CONFIG)
    update_form_data_headers = dingtalkyida__1__0_models.UpdateFormDataHeaders()
    update_form_data_headers.x_acs_dingtalk_access_token = get_token()
    update_form_data_request = dingtalkyida__1__0_models.UpdateFormDataRequest(
        system_token=SYSTOKEN,
        app_type=APP,
        form_instance_id=fiid,
        user_id=UID,
        update_form_data_json=dumps(data)
    )
    try:
        client.update_form_data_with_options(update_form_data_request, update_form_data_headers, util_models.RuntimeOptions())
    except Exception as err:
        raise Exception(f"更新{name}数据错误（{err.code}）： {err.message}")


def search(condition: dict | list[dict], page_number: int, form: str, name: str = ""):
    client = dingtalkyida_1_0Client(CONFIG)
    search_form_data_second_generation_no_table_field_headers = (
        dingtalkyida__1__0_models.SearchFormDataSecondGenerationNoTableFieldHeaders()
    )
    search_form_data_second_generation_no_table_field_headers.x_acs_dingtalk_access_token = (
        get_token()
    )
    search_form_data_second_generation_no_table_field_request = (
        dingtalkyida__1__0_models.SearchFormDataSecondGenerationNoTableFieldRequest(
            system_token=SYSTOKEN,
            page_size=100,
            page_number=page_number,
            form_uuid=form,
            user_id=UID,
            app_type=APP,
            search_condition=dumps(condition),
        )
    )

    try:
        return client.search_form_data_second_generation_no_table_field_with_options(
            search_form_data_second_generation_no_table_field_request,
            search_form_data_second_generation_no_table_field_headers,
            util_models.RuntimeOptions(),
        )
    except Exception as err:
        raise Exception(f"获取{name}数据错误（{err.code}）： {err.message}")


@retry(10, 2)
def ding_push(__str: str, __type: str):
    # 获取DING消息对象
    page_number = 1
    uuids = []
    users = []
    condition = [
        {
            "key": "radioField_lya3eeaa",
            "value": "DING提醒",
            "type": "ARRAY",
            "operator": "eq",
            "componentName": "RadioField",
        },
        {
            "key": "multiSelectField_ly886nbk",
            "value": [__type],
            "type": "ARRAY",
            "operator": "contains",
            "componentName": "MultiSelectField",
        },
    ]

    while True:
        response = search(condition, page_number, SUBS, "DING订阅用户")
        for i in response.body.data:
            uuids.extend(i.form_data["employeeField_lxug4j9k_id"])
            users.extend(i.form_data["employeeField_lxug4j9k"])

        if response.body.total_count > page_number * 100:
            page_number += 1
            sleep(1)
        else:
            break

    # DING消息发送
    if uuids:
        client = dingtalkrobot_1_0Client(CONFIG)
        robot_send_ding_headers = dingtalkrobot__1__0_models.RobotSendDingHeaders()
        robot_send_ding_headers.x_acs_dingtalk_access_token = get_token()
        robot_send_ding_request = dingtalkrobot__1__0_models.RobotSendDingRequest(
            robot_code="ding2gb8oi5iswe8mj7q",
            remind_type=1,
            receiver_user_id_list=uuids,
            content=__str,
        )
        try:
            client.robot_send_ding_with_options(
                robot_send_ding_request,
                robot_send_ding_headers,
                util_models.RuntimeOptions(),
            )
            update_log("发DING成功，{}告警，接收人：{}".format(__type, "、".join(users)))
        except Exception as err:
            raise Exception(f"发DING失败 ({err.code}): {err.message}")
    return 0


@retry(10, 2)
def running_push(__dict: dict, __timestamp: int):
    page_number = 1
    fiid = []
    condition = [{
        "key": "dateField_lz714eu6",
        "value": int(1000 * time()),
        "type": "DOUBLE",
        "operator": "gt",
        "componentName": "DateField",
    }]

    while True:
        response = search(condition, page_number, RUNSUB, "运行态势订阅")
        for i in response.body.data:
            fiid.append(i.form_instance_id)

        if response.body.total_count > page_number * 100:
            page_number += 1
            sleep(1)
        else:
            break

    table = TABLE.copy()
    table[RENAMER_RUNSUB["时间"]] = __timestamp
    for k, v in RENAMER_RUNSUB.items():
        if vv := __dict.get(k):
            table[v]["data"] = vv
    for fiid in fiid:
        update(table, fiid, "运行态势")


def update_code(_t: float):
    t1, t2 = _t % 1800000, _t % 3600000
    return (
        2 if t2 < 180000 or t2 >= 3480000 else 1 if t1 < 180000 or t1 >= 1680000 else 0
    )


if __name__ == "__main__":
    update_log("进程启动")
    updated_info = updated_warn = updated_duty = duty_mtime = 0
    retries_info = retries_warn = retries_duty = 5
    while True:
        try:
            json = f"{PATH}/sync.json"
            if os.path.exists(json):
                with open(json) as json:
                    json_info = load(json)
            else:
                json_info = {}
            json = f"{PATH}/moni.json"
            if os.path.exists(json):
                with open(json) as json:
                    json_warn = load(json)
            else:
                json_warn = {}
            json = f"{PATH}/duty.json"
            if os.path.exists(json):
                duty_mtime = os.path.getmtime(json)
                with open(json) as json:
                    json_duty = load(json)
            else:
                json_duty = []
            assert isinstance(json_warn, dict) and isinstance(json_info, dict) and isinstance(json_duty, list)
        except KeyboardInterrupt:
            break
        except Exception as exception:
            exception = repr(exception)
            update_log(f"加载JSON失败 ({exception[: exception.find('(')]})，重试")
            sleep(1)
            continue

        try:
            if retries_info <= 0:
                updated_info = json_info.get("数据日期时间", 0)
                update_log("信息同步尝试5次均失败，跳过")
            elif json_info.get("数据日期时间", 0) != updated_info:
                upload_json = dict()
                for k, v in RENAMER_INFO.items():
                    vv = json_info.get(k, "")
                    if "图片" in k and vv.endswith(".png"):
                        if vv := upload_img(vv):
                            pass
                        else:
                            continue
                    elif isinstance(vv, str):
                        if vv.endswith("%"):
                            vv = float(vv[:-1]) * 0.01
                        elif not vv:
                            continue
                    upload_json[v] = vv
                upload_json["numberField_lxvgxw12"] = update_code(json_info["数据日期时间"])
                upload([dumps(upload_json)], INFO, "信息")

                updated_info = json_info["数据日期时间"]
                update_log("信息同步成功")
            retries_info = 5
        except KeyboardInterrupt:
            break
        except Exception:
            update_log(format_exc().split("\n", 1)[1])
            sleep(1)
            TOKEN_TIMESTAMP = 0
            retries_info -= 1

        try:
            if retries_warn <= 0:
                updated_warn = json_warn.get("数据日期时间", 0)
                update_log("告警同步尝试5次均失败，跳过")
            elif json_warn.get("数据日期时间", 0) != updated_warn:
                if v := json_warn.get("态势"):
                    running_push(v, json_warn["数据日期时间"])
                    sleep(1)
                    type_min = datetime.now().minute % 10
                    vv = []
                    for k, v in v.items():
                        if i := TABLE.get(RENAMER_RUNSUB.get(k), {}).copy():
                            i["data"] = v
                            i = {
                                RENAMER_RUNNING["类型"]: type_min,
                                RENAMER_RUNNING["标题"]: k,
                                RENAMER_RUNNING["时间"]: json_warn["数据日期时间"],
                                RENAMER_RUNNING["表格"]: i,
                            }
                            vv.append(dumps(i))
                    upload(vv, RUNNING, "态势")
                    update_log("态势同步成功")

                if v := json_warn.get("告警", []):
                    warns, warnlist = dict(), []
                    for v in v:
                        if v["类型"] != "解除":
                            vv = ""
                            if v["登机门"]:
                                vv += f"/登机口{v['登机门']}"
                            if v["机位"]:
                                vv += f"/机位{v['机位']}"
                            if v["地服"]:
                                vv += f"/{v['地服']}"
                            if v["目的地"]:
                                vv += f"/{v['目的地']}"
                            warns.setdefault(v["类型"], []).append(
                                "{}/{} {:%H%M}{}：{}".format(
                                    v["航班号"],
                                    "STA" if "进港" in v["类型"] or "落地" in v["类型"] else "STD",
                                    datetime.fromtimestamp(v["日期"] // 1000),
                                    vv,
                                    v["描述"],
                                )
                            )
                        warnlist.append(dumps(dict((RENAMER_WARN[k], v) for k, v in v.items() if v and k in RENAMER_WARN)))
                    sleep(1)
                    upload(warnlist, WARN, "告警")
                    update_log("告警同步成功")

                    for k, v in warns.items():
                        sleep(1)
                        ding_push("；\n".join(v) + "。", k)

                updated_warn = json_warn["数据日期时间"]
            retries_warn = 5
        except KeyboardInterrupt:
            break
        except Exception:
            update_log(format_exc().split("\n", 1)[1])
            sleep(1)
            TOKEN_TIMESTAMP = 0
            retries_warn -= 1

        try:
            if retries_duty <= 0:
                updated_duty = os.path.getmtime()
                update_log("值班同步尝试5次均失败，跳过")
            elif duty_mtime != updated_duty:
                upload([dumps(i) for i in json_duty], DUTY, "值班", "APP_HZ4NPKJBV8VV04401NO7", "3L966J711A24UKS1BGQVOB579DD32OF9BK98LC2")
                updated_duty = duty_mtime
                update_log("值班同步成功")
            retries_duty = 5
        except KeyboardInterrupt:
            break
        except Exception:
            update_log(format_exc().split("\n", 1)[1])
            TOKEN_TIMESTAMP = 0
            retries_duty -= 1

        clear_log()
        sleep(5)
