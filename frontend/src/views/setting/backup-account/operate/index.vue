<template>
    <DrawerPro v-model="drawerVisible" :header="title + $t('setting.backupAccount')" @close="handleClose" size="large">
        <el-form @submit.prevent ref="formRef" v-loading="loading" label-position="top" :model="dialogData.rowData">
            <el-form-item :label="$t('commons.table.name')" prop="name" :rules="Rules.requiredInput">
                <el-tag v-if="dialogData.title === 'edit'">
                    {{ dialogData.rowData!.name === 'localhost' ? $t('terminal.local') : dialogData.rowData!.name }}
                </el-tag>
                <el-input v-else v-model="dialogData.rowData!.name" />
            </el-form-item>
            <el-form-item :label="$t('setting.scope')" prop="isPublic" :rules="Rules.requiredSelect">
                <el-tag v-if="dialogData.title === 'edit'">
                    {{ dialogData.rowData!.isPublic ? $t('setting.public') : $t('setting.private') }}
                </el-tag>
                <el-radio-group v-else v-model="dialogData.rowData!.isPublic">
                    <el-radio :value="true" size="large">{{ $t('setting.public') }}</el-radio>
                    <el-radio :value="false" size="large">{{ $t('setting.private') }}</el-radio>
                </el-radio-group>
                <span class="input-help">
                    {{ dialogData.rowData!.isPublic ? $t('setting.publicHelper') : $t('setting.privateHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                <el-tag v-if="dialogData.title === 'edit'">{{ $t('setting.' + dialogData.rowData!.type) }}</el-tag>
                <el-select v-else v-model="dialogData.rowData!.type" @change="changeType">
                    <el-option :label="$t('setting.S3')" value="S3"></el-option>
                    <el-option :label="$t('setting.MINIO')" value="MINIO"></el-option>
                    <el-option :label="$t('setting.SFTP')" value="SFTP"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                :label="$t('setting.mode')"
                prop="varsJson.mode"
                :rules="Rules.requiredSelect"
            >
                <el-radio-group v-model="dialogData.rowData!.varsJson['mode']">
                    <el-radio value="virtual hosted">Virtual Hosted</el-radio>
                    <el-radio value="path">Path</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="hasAccessKey()" label="Access Key ID" prop="accessKey" :rules="Rules.requiredInput">
                <el-input clearable v-model.trim="dialogData.rowData!.accessKey" />
            </el-form-item>
            <el-form-item v-if="hasAccessKey()" label="Secret Key" prop="credential" :rules="Rules.requiredInput">
                <el-input show-password clearable v-model.trim="dialogData.rowData!.credential" />
            </el-form-item>
            <div v-if="dialogData.rowData!.type === 'SFTP'">
                <el-form-item :label="$t('setting.address')" prop="varsJson.address" :rules="Rules.host">
                    <el-input v-model.trim="dialogData.rowData!.varsJson['address']" clearable />
                </el-form-item>
                <el-form-item :label="$t('commons.table.port')" prop="varsJson.port" :rules="[Rules.port]">
                    <el-input-number :min="0" :max="65535" v-model.number="dialogData.rowData!.varsJson['port']" />
                </el-form-item>
            </div>
            <div v-if="hasPassword()">
                <el-form-item :label="$t('commons.login.username')" prop="accessKey" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="dialogData.rowData!.accessKey" />
                </el-form-item>

                <div v-if="dialogData.rowData!.type === 'SFTP'">
                    <el-form-item :label="$t('terminal.authMode')" prop="varsJson.authMode">
                        <el-radio-group v-model="dialogData.rowData!.varsJson['authMode']">
                            <el-radio value="password">{{ $t('terminal.passwordMode') }}</el-radio>
                            <el-radio value="key">{{ $t('terminal.keyMode') }}</el-radio>
                        </el-radio-group>
                    </el-form-item>
                </div>
                <div v-if="dialogData.rowData!.type === 'SFTP' && dialogData.rowData!.varsJson['authMode'] === 'key'">
                    <el-form-item :label="$t('terminal.key')" prop="credential" :rules="[Rules.requiredInput]">
                        <el-input type="textarea" v-model="dialogData.rowData!.credential" />
                    </el-form-item>
                    <el-form-item :label="$t('terminal.keyPassword')" prop="varsJson.passPhrase">
                        <el-input
                            type="password"
                            show-password
                            clearable
                            v-model="dialogData.rowData!.varsJson['passPhrase']"
                        />
                    </el-form-item>
                </div>
                <el-form-item
                    v-else
                    :label="$t('commons.login.password')"
                    prop="credential"
                    :rules="[Rules.requiredInput]"
                >
                    <el-input type="password" clearable show-password v-model.trim="dialogData.rowData!.credential" />
                </el-form-item>
            </div>
            <el-form-item v-if="hasRemember()" prop="rememberAuth">
                <el-checkbox v-model="dialogData.rowData!.rememberAuth">
                    {{ $t('terminal.rememberPassword') }}
                </el-checkbox>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                label="Region"
                prop="varsJson.region"
                :rules="Rules.requiredInput"
            >
                <el-input v-model.trim="dialogData.rowData!.varsJson['region']" />
            </el-form-item>
            <el-form-item
                v-if="hasAccessKey()"
                label="Endpoint"
                prop="varsJson.endpointItem"
                :rules="Rules.requiredInput"
            >
                <el-input v-model.trim="dialogData.rowData!.varsJson['endpointItem']">
                    <template #prepend>
                        <el-select v-model.trim="domainProto" class="p-w-100">
                            <el-option label="http" value="http" />
                            <el-option label="https" value="https" />
                        </el-select>
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item v-if="hasAccessKey()" label="Bucket" prop="bucket" :rules="Rules.requiredInput">
                <el-checkbox v-model="dialogData.rowData!.bucketInput" :label="$t('container.input')" />
                <el-input clearable v-if="dialogData.rowData!.bucketInput" v-model="dialogData.rowData!.bucket" />
                <div v-else class="w-full">
                    <el-select class="!w-4/5" v-model="dialogData.rowData!.bucket">
                        <el-option v-for="item in buckets" :key="item" :value="item" />
                    </el-select>
                    <el-button class="!w-1/5" plain @click="getBuckets(formRef)">
                        {{ $t('setting.loadBucket') }}
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'S3'"
                :label="$t('setting.scType')"
                prop="varsJson.scType"
                :rules="[Rules.requiredSelect]"
            >
                <el-select v-model="dialogData.rowData!.varsJson['scType']">
                    <el-option value="STANDARD" :label="$t('setting.scStandard')" />
                    <el-option value="STANDARD_IA" :label="$t('setting.scStandard_IA')" />
                    <el-option value="GLACIER" :label="$t('setting.scArchive')" />
                    <el-option value="DEEP_ARCHIVE" :label="$t('setting.scDeep_Archive')" />
                </el-select>
                <el-alert
                    v-if="dialogData.rowData!.varsJson['scType'] === 'GLACIER' || dialogData.rowData!.varsJson['scType'] === 'DEEP_ARCHIVE'"
                    class="mt-2.5"
                    :closable="false"
                    type="warning"
                    :title="$t('setting.archiveHelper')"
                />
            </el-form-item>
            <el-form-item v-if="hasBackDir()" :label="$t('setting.backupDir')" prop="backupPath">
                <el-input clearable v-model.trim="dialogData.rowData!.backupPath" placeholder="/1panel" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'SFTP'"
                :label="$t('setting.backupDir')"
                prop="backupPath"
                :rules="[Rules.requiredInput]"
            >
                <el-input v-model.trim="dialogData.rowData!.backupPath" />
            </el-form-item>
            <el-form-item
                v-if="dialogData.rowData!.type === 'LOCAL'"
                :label="$t('setting.backupDir')"
                prop="backupPath"
                :rules="Rules.requiredInput"
            >
                <el-input v-model="dialogData.rowData!.backupPath">
                    <template #prepend>
                        <el-button icon="Folder" @click="fileRef.acceptParams({ dir: true })" />
                    </template>
                </el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="handleClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button :disabled="loading" @click="onCheck(formRef)">
                {{ $t('terminal.testConn') }}
            </el-button>
            <el-button type="primary" :disabled="!isOK || loading" @click="onSubmit()">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <FileList ref="fileRef" @choose="loadDir" />
</template>

<script lang="ts" setup>
import { ref, watch, computed, onUnmounted } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import FileList from '@/components/file-list/index.vue';
import { addBackup, checkBackup, editBackup, listBucket } from '@/api/modules/backup';
import { deepCopy, spliceHttp, splitHttp } from '@/utils/util';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Base64 } from 'js-base64';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const buckets = ref();
const fileRef = ref();

const isOK = ref();
const stopWatch = ref();

const domainProto = ref('http');
const emit = defineEmits(['search']);

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const formWatcher = computed(() => {
    const { type, isPublic, accessKey, bucket, credential, backupPath, bucketInput, varsJson } =
        dialogData.value.rowData || {};
    return { type, isPublic, accessKey, bucket, credential, backupPath, bucketInput, varsJson };
});
const startWatcher = () => {
    if (stopWatch.value) {
        stopWatcher();
    }
    stopWatch.value = watch(
        () => formWatcher.value,
        () => {
            stopWatcher();
            isOK.value = false;
        },
        { deep: true },
    );
};
const stopWatcher = () => {
    if (stopWatch.value) {
        stopWatch.value();
        stopWatch.value = null;
    }
};

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    dialogData.value.rowData.varsJson = dialogData.value.rowData!.vars
        ? JSON.parse(dialogData.value.rowData!.vars)
        : {};
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
    if (dialogData.value.title === 'create') {
        dialogData.value.rowData!.type = 'S3';
        changeType();
        drawerVisible.value = true;
        return;
    }
    buckets.value = [];
    if (hasAccessKey()) {
        let itemJson = dialogData.value.rowData!.varsJson['endpoint'];
        let httpItem = splitHttp(itemJson);
        dialogData.value.rowData!.varsJson['endpointItem'] = httpItem.url;
        domainProto.value = httpItem.proto;
    }
    if (dialogData.value.rowData!.rememberAuth) {
        dialogData.value.rowData!.accessKey = Base64.decode(dialogData.value.rowData!.accessKey);
        dialogData.value.rowData!.credential = Base64.decode(dialogData.value.rowData!.credential);
    }
    drawerVisible.value = true;
};
const hasRemember = () => {
    return dialogData.value.rowData!.type !== 'LOCAL';
};
const hasAccessKey = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'MINIO' || itemType === 'S3';
};
const hasPassword = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType === 'SFTP';
};

const hasBackDir = () => {
    let itemType = dialogData.value.rowData!.type;
    return itemType !== 'LOCAL' && itemType !== 'SFTP';
};

const loadDir = async (path: string) => {
    dialogData.value.rowData!.backupPath = path;
};

const changeType = () => {
    buckets.value = [];
    dialogData.value.rowData!.varsJson = {};
    dialogData.value.rowData!.rememberAuth = false;
    switch (dialogData.value.rowData!.type) {
        case 'S3':
            dialogData.value.rowData.varsJson['scType'] = 'STANDARD';
            dialogData.value.rowData.varsJson['mode'] = 'virtual hosted';
            break;
        case 'SFTP':
            dialogData.value.rowData.varsJson['port'] = 22;
            dialogData.value.rowData.varsJson['authMode'] = 'password';
            break;
    }
};

function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

const getBuckets = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result1 = await formEl.validateField('varsJson.endpointItem', callback);
    const result2 = await formEl.validateField('accessKey', callback);
    const result3 = await formEl.validateField('credential', callback);
    const result4 = await formEl.validateField('varsJson.region', callback);
    if (!result1 || !result2 || !result3 || !result4) {
        return;
    }
    loading.value = true;
    let item = deepCopy(dialogData.value.rowData!.varsJson);
    let itemEndpoint = loadEndpoint();
    item['endpoint'] = itemEndpoint;
    item['endpointItem'] = undefined;
    listBucket({
        isPublic: dialogData.value.rowData!.isPublic,
        type: dialogData.value.rowData!.type,
        vars: JSON.stringify(item),
        accessKey: dialogData.value.rowData!.accessKey,
        credential: dialogData.value.rowData!.credential,
    })
        .then((res) => {
            loading.value = false;
            buckets.value = res.data;
        })
        .catch(() => {
            buckets.value = [];
            loading.value = false;
        });
};

const loadEndpoint = () => {
    let item = splitHttp(dialogData.value.rowData!.varsJson['endpointItem']);
    if (item.proto) {
        domainProto.value = item.proto;
        dialogData.value.rowData!.varsJson['endpointItem'] = item.url;
    }
    return spliceHttp(domainProto.value, dialogData.value.rowData!.varsJson['endpointItem']);
};

const onCheck = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!dialogData.value.rowData) return;
        if (hasAccessKey()) {
            let itemEndpoint = loadEndpoint();
            dialogData.value.rowData!.varsJson['endpoint'] = itemEndpoint;
        }
        dialogData.value.rowData.vars = JSON.stringify(dialogData.value.rowData!.varsJson);
        loading.value = true;
        await checkBackup(dialogData.value.rowData)
            .then((res) => {
                loading.value = false;
                if (res.data.isOk) {
                    isOK.value = true;
                    MsgSuccess(i18n.global.t('terminal.connTestOk'));
                    startWatcher();
                    return;
                }
                isOK.value = false;
                MsgError(i18n.global.t('terminal.connTestFailed') + ':' + res.data.msg);
            })
            .catch(() => {
                loading.value = false;
                isOK.value = false;
            });
    });
};

const onSubmit = async () => {
    dialogData.value.rowData.vars = JSON.stringify(dialogData.value.rowData!.varsJson);
    loading.value = true;
    if (dialogData.value.title === 'create') {
        await addBackup(dialogData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                drawerVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
        return;
    }
    await editBackup(dialogData.value.rowData)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            drawerVisible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

onUnmounted(() => {
    if (stopWatch.value) stopWatcher();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.option-help {
    float: right;
    font-size: 12px;
    word-break: break-all;
    color: #8f959e;
}
</style>
