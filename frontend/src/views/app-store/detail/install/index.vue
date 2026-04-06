<template>
    <DrawerPro v-model="open" :header="$t('commons.button.install')" @close="handleClose" size="large">
        <template #content>
            <div v-if="requiresOfflinePackage && preflight" class="mb-4">
                <el-alert
                    :title="preflight.ready ? '离线安装环境已就绪' : '离线安装环境未完全就绪'"
                    :type="preflight.ready ? 'success' : 'warning'"
                    :closable="false"
                    show-icon
                />
                <div class="preflight-grid mt-3">
                    <el-alert
                        :title="preflight.dockerMessage"
                        :type="preflight.dockerReady ? 'success' : 'error'"
                        :closable="false"
                        show-icon
                    />
                    <el-alert
                        :title="preflight.imageMessage"
                        :type="preflight.imageReady ? 'success' : 'warning'"
                        :closable="false"
                        show-icon
                    />
                    <el-alert
                        :title="preflight.installDirMessage"
                        :type="preflight.installDirReady ? 'success' : 'error'"
                        :closable="false"
                        show-icon
                    />
                    <el-alert
                        :title="preflight.diskMessage"
                        :type="preflight.diskReady ? 'success' : 'warning'"
                        :closable="false"
                        show-icon
                    />
                    <el-alert
                        v-if="appKey === 'openresty'"
                        :title="preflight.websiteDirMessage"
                        :type="preflight.websiteDirReady ? 'success' : 'error'"
                        :closable="false"
                        show-icon
                    />
                </div>
                <div class="offline-upload mt-3">
                    <el-alert
                        :title="imageReady ? `镜像已就绪：${requiredImage}` : `缺少离线镜像：${requiredImage}`"
                        :type="imageReady ? 'success' : 'warning'"
                        :closable="false"
                        show-icon
                    />
                    <el-alert
                        v-if="!imageReady"
                        title="当前应用尚未检测到所需镜像，请先导入镜像 tar 包，或直接在此上传。"
                        type="info"
                        :closable="false"
                        show-icon
                        class="mt-3"
                    />
                    <el-upload
                        v-if="requiresOfflinePackage"
                        class="offline-uploader mt-3"
                        drag
                        :auto-upload="false"
                        :limit="1"
                        :on-change="handleFileChange"
                        :on-remove="handleFileRemove"
                        :file-list="fileList"
                    >
                        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                        <div class="el-upload__text">{{ uploadHint }}</div>
                    </el-upload>
                </div>
            </div>
            <AppInstallForm
                ref="installFormRef"
                v-model="formData"
                :loading="loading"
                :batch-install-support="batchInstallSupport"
            />
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="handleSubmit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <TaskLog ref="taskLogRef" />
</template>

<script lang="ts" setup name="AppInstallPage">
import { computed, nextTick, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import type { UploadFile, UploadFiles } from 'element-plus';
import { UploadFilled } from '@element-plus/icons-vue';
import AppInstallForm from '@/views/app-store/detail/form/index.vue';
import {
    installApp,
    installDatabaseApp,
    installOpenrestyApp,
    installAppToNodes,
    preflightOfflineInstall,
    uploadDatabaseImage,
    uploadOpenrestyImage,
} from '@/api/modules/app';
import { listAllImage } from '@/api/modules/container';
import { MsgError, MsgWarning } from '@/utils/message';
import { newUUID } from '@/utils/util';
import { routerToName } from '@/utils/router';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';

const router = useRouter();
const open = ref(false);
const loading = ref(false);
const installFormRef = ref<InstanceType<typeof AppInstallForm>>();
const taskLogRef = ref();
const appKey = ref('');
const batchInstallSupport = ref(false);
const uploadFile = ref<File>();
const fileList = ref<UploadFiles>([]);
const imageTags = ref<string[]>([]);
const preflight = ref();

const builtInDatabaseApps = ['mysql', 'postgresql', 'redis'];
const builtInImageMap: Record<string, string> = {
    mysql: 'mysql:8.0',
    postgresql: 'postgres:16',
    redis: 'redis:7',
    openresty: '1panel/openresty:1.27.1.2-3-3-focal',
};

const requiresOfflinePackage = computed(() => {
    return builtInDatabaseApps.includes(appKey.value) || appKey.value === 'openresty';
});
const requiredImage = computed(() => builtInImageMap[appKey.value] || '');
const imageReady = computed(() => {
    if (preflight.value && requiredImage.value) {
        return preflight.value.imageReady;
    }
    if (!requiredImage.value) {
        return true;
    }
    return imageTags.value.includes(requiredImage.value);
});
const uploadHint = computed(() => {
    if (appKey.value === 'openresty') {
        return '上传 OpenResty 离线镜像 tar 包';
    }
    if (appKey.value) {
        return `上传 ${appKey.value} 离线镜像 tar 包`;
    }
    return '上传离线镜像 tar 包';
});

const formData = reactive({
    appDetailId: 0,
    params: {},
    name: '',
    advanced: true,
    cpuQuota: 0,
    memoryLimit: 0,
    memoryUnit: 'M',
    containerName: '',
    allowPort: false,
    editCompose: false,
    dockerCompose: '',
    version: '',
    appID: '',
    pullImage: false,
    taskID: '',
    gpuConfig: false,
    specifyIP: '',
    format: 'utf8mb4',
    collation: '',
    restartPolicy: 'always',
    pushNode: false,
    nodes: [],
});

const resetUploadState = () => {
    uploadFile.value = undefined;
    fileList.value = [];
};

const handleFileChange = (file: UploadFile, files: UploadFiles) => {
    fileList.value = files.slice(-1);
    uploadFile.value = file.raw;
};

const handleFileRemove = () => {
    resetUploadState();
};

const loadImageState = async () => {
    const res = await listAllImage();
    const tags: string[] = [];
    for (const item of res.data || []) {
        for (const tag of item.tags || []) {
            if (tag && tag !== '<none>:<none>') {
                tags.push(tag);
            }
        }
    }
    imageTags.value = tags;
};

const loadPreflight = async () => {
    if (!requiresOfflinePackage.value || !appKey.value) {
        preflight.value = undefined;
        return;
    }
    const res = await preflightOfflineInstall(appKey.value);
    preflight.value = res.data;
};

const handleClose = () => {
    open.value = false;
    resetUploadState();
    preflight.value = undefined;
    installFormRef.value?.resetForm();
    if (router.currentRoute.value.query.install) {
        routerToName('AppAll');
    }
};

const uploadImageIfNeeded = async () => {
    if (!requiresOfflinePackage.value) {
        return '';
    }
    if (!uploadFile.value) {
        if (imageReady.value) {
            return '';
        }
        MsgWarning('请先上传离线镜像 tar 包');
        throw new Error('missing offline package');
    }
    if (builtInDatabaseApps.includes(appKey.value)) {
        const res = await uploadDatabaseImage(appKey.value, uploadFile.value);
        await loadImageState();
        await loadPreflight();
        return res.data;
    }
    const res = await uploadOpenrestyImage(uploadFile.value);
    await loadImageState();
    await loadPreflight();
    return res.data;
};

const submitBuiltInInstall = async (taskID: string, imagePath: string, submitData: any) => {
    const params = submitData?.params || {};

    if (!imageReady.value && !imagePath) {
        MsgWarning(`请先导入所需镜像：${requiredImage.value}`);
        throw new Error('image not ready');
    }

    if (builtInDatabaseApps.includes(appKey.value)) {
        const passwordKey = appKey.value === 'redis' ? 'PANEL_REDIS_ROOT_PASSWORD' : 'PANEL_DB_ROOT_PASSWORD';
        await installDatabaseApp({
            appKey: appKey.value,
            port: Number(params.PANEL_APP_PORT_HTTP),
            password: params[passwordKey],
            imagePath,
            taskID,
            containerName: submitData.containerName || '',
            installDir: '',
        });
        return;
    }

    await installOpenrestyApp({
        httpPort: Number(params.PANEL_APP_PORT_HTTP),
        httpsPort: Number(params.PANEL_APP_PORT_HTTPS),
        imagePath,
        taskID,
        containerName: submitData.containerName || '',
        websiteDir: params.WEBSITE_DIR || '',
        packageUrl: '',
    });
};

const install = async (submitData: any) => {
    loading.value = true;
    const taskID = newUUID();
    submitData.taskID = taskID;

    try {
        const imagePath = await uploadImageIfNeeded();
        if (submitData.pushNode) {
            submitData.appKey = appKey.value;
            await installAppToNodes(submitData);
        } else if (requiresOfflinePackage.value) {
            await submitBuiltInInstall(taskID, imagePath, submitData);
        } else {
            await installApp(submitData);
        }
        handleClose();
        openTaskLog(taskID);
    } finally {
        loading.value = false;
    }
};

const handleSubmit = async () => {
    const isValid = await installFormRef.value?.validate();
    if (!isValid) return;

    const submitData = installFormRef.value?.getFormData();

    if (submitData.editCompose && submitData.dockerCompose === '') {
        MsgError(i18n.global.t('app.composeNullErr'));
        return;
    }

    if (submitData.cpuQuota < 0) {
        submitData.cpuQuota = 0;
    }
    if (submitData.memoryLimit < 0) {
        submitData.memoryLimit = 0;
    }

    if (preflight.value && !preflight.value.ready && !(requiresOfflinePackage.value && !preflight.value.imageReady)) {
        MsgWarning('请先完成离线安装前置检查');
        return;
    }

    const isHostMode = installFormRef.value?.isHostMode();
    if (!isHostMode && !submitData.allowPort) {
        ElMessageBox.confirm(i18n.global.t('app.installWarn'), i18n.global.t('app.checkTitle'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(async () => {
            await install(submitData);
        });
    } else {
        await install(submitData);
    }
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const acceptParams = async (props: { app: any; params?: any }) => {
    appKey.value = props.app.key;
    batchInstallSupport.value = props.app.batchInstallSupport;
    open.value = true;
    await nextTick();
    resetUploadState();
    installFormRef.value?.resetForm();
    installFormRef.value?.initForm(props.app.key);
    await loadImageState();
    await loadPreflight();
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.preflight-grid {
    display: grid;
    gap: 12px;
}

.offline-uploader {
    width: 100%;
}
</style>
