<template>
    <LayoutContent v-loading="loading" :title="activeName">
        <template #search>
            <Tags @change="changeTag" hideKey="Runtime" />
        </template>
        <template #leftToolBar>
            <el-button @click="sync" type="primary" plain v-if="mode === 'installed' && data != null">
                {{ $t('commons.button.refresh') }}
            </el-button>
        </template>
        <template #rightToolBar>
            <TableSearch @search="search" v-model:searchName="searchReq.name" />
        </template>
        <template #main>
            <div>
                <MainDiv :heightDiff="mode === 'upgrade' ? 280 : 300">
                    <div class="update-prompt" v-if="data === null || data.length === 0">
                        <span>{{ mode === 'upgrade' ? $t('app.updatePrompt') : $t('app.installPrompt') }}</span>
                        <div>
                            <img src="@/assets/images/no_update_app.svg" />
                        </div>
                    </div>
                    <el-row :gutter="5" v-else>
                        <el-col
                            v-for="installed in data"
                            :key="installed.id"
                            :xs="24"
                            :sm="24"
                            :md="24"
                            :lg="12"
                            :xl="12"
                        >
                            <AppCard
                                :installed="installed"
                                :mode="mode"
                                :defaultLink="defaultLink"
                                :currentNode="currentNode"
                                @open-detail="openDetail(installed.appKey)"
                                @open-backups="openBackups(installed)"
                                @open-log="openLog(installed)"
                                @open-terminal="openTerminal(installed)"
                                @open-operate="openOperate(installed, 'upgrade')"
                                @favorite-install="favoriteInstall(installed)"
                                @to-folder="toFolder(installed)"
                                @open-uploads="openUploads(installed)"
                                @jump-to-path="jumpToPath(router, '/settings/panel')"
                                @to-container="toContainer(installed)"
                                @ignore-app="openOperate(installed, 'ignore')"
                            >
                                <template #buttons>
                                    <div
                                        class="d-button flex flex-wrap items-center justify-start gap-1.5"
                                        v-if="mode === 'installed' && installed.status != 'Installing'"
                                    >
                                        <el-button
                                            class="app-button"
                                            plain
                                            round
                                            size="small"
                                            @click="operate(installed, 'restart')"
                                            :disabled="installed.status !== 'Running'"
                                        >
                                            {{ $t('commons.operate.restart') }}
                                        </el-button>
                                        <el-button
                                            class="app-button"
                                            plain
                                            round
                                            size="small"
                                            @click="uninstall(installed)"
                                            :disabled="installed.status === 'Upgrading'"
                                        >
                                            {{ $t('commons.button.uninstall') }}
                                        </el-button>
                                    </div>
                                </template>
                            </AppCard>
                        </el-col>
                    </el-row>
                </MainDiv>
            </div>
            <div class="page-button">
                <fu-table-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    v-bind="paginationConfig"
                    @change="search"
                    :layout="'total, sizes, prev, pager, next, jumper'"
                />
            </div>
        </template>
    </LayoutContent>
    <TaskLog ref="taskLogRef" @close="search" />
    <Detail ref="detailRef" />
    <AppUpgrade ref="upgradeRef" @close="search" />
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessageBox } from 'element-plus';
import { v4 as uuidv4 } from 'uuid';
import AppCard from '@/views/app-store/installed/app/card.vue';
import TaskLog from '@/components/log/task/index.vue';
import Detail from '@/views/app-store/detail/index.vue';
import Tags from '@/views/app-store/components/tag.vue';
import MainDiv from '@/components/main-div/index.vue';
import AppUpgrade from '@/views/app-store/installed/upgrade/index.vue';
import { appInstalledDeleteCheck, installedOp, searchAppInstalled, syncInstalledApp } from '@/api/modules/app';
import { jumpToPath } from '@/utils/util';
import { routerToFileWithPath, routerToNameWithQuery } from '@/utils/router';
import i18n from '@/lang';
import { getAgentSettingByKey } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { currentNode, isMaster, currentNodeAddr } = useGlobalStore();

const router = useRouter();
const data = ref<any[] | null>(null);
const loading = ref(false);
const paginationConfig = reactive({
    cacheSizeKey: 'app-installed-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('app-installed-page-size')) || 20,
    total: 0,
});
const searchReq = reactive({
    page: 1,
    pageSize: 20,
    name: '',
    tags: [],
    update: false,
    sync: false,
});
const activeName = ref(i18n.global.t('app.installed'));
const mode = ref('installed');
const defaultLink = ref('');
const detailRef = ref();
const taskLogRef = ref();
const upgradeRef = ref();

const openDetail = (key: string) => {
    detailRef.value.acceptParams(key, 'install');
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID, true, currentNode.value);
};

const changeTag = (key: string) => {
    searchReq.tags = [];
    if (key !== 'all') {
        searchReq.tags = [key];
    }
    search();
};

const search = async () => {
    searchReq.page = paginationConfig.currentPage;
    searchReq.pageSize = paginationConfig.pageSize;
    localStorage.setItem('app-installed-page-size', String(searchReq.pageSize));
    loading.value = true;
    try {
        const res = await searchAppInstalled(searchReq);
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
    } finally {
        loading.value = false;
    }
};

const sync = async () => {
    if (mode.value === 'installed') {
        await syncInstalledApp();
    }
    await search();
};

const openOperate = (row: any, op: string) => {
    if (op === 'upgrade' || op === 'ignore') {
        upgradeRef.value.acceptParams(row, op, currentNode.value);
    }
};

const operate = async (row: any, op: string) => {
    const taskID = uuidv4();
    await installedOp({ installId: row.id, operate: op, taskID }, currentNode.value);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    openTaskLog(taskID);
};

const uninstall = async (row: any) => {
    const res = await appInstalledDeleteCheck(row.id, currentNode.value);
    const resourceMsg = (res.data || []).map((item: any) => `${item.type}: ${item.name}`).join('\n');
    const message = resourceMsg ? `${i18n.global.t('app.deleteHelper')}\n${resourceMsg}` : i18n.global.t('app.deleteHelper');
    await ElMessageBox.confirm(message, i18n.global.t('commons.button.uninstall'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    });
    const taskID = uuidv4();
    await installedOp({ installId: row.id, operate: 'delete', taskID }, currentNode.value);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    openTaskLog(taskID);
};

const toFolder = async (row: any) => {
    if (!row.path) {
        return;
    }
    await routerToFileWithPath(row.path);
};

const openTerminal = async (row: any) => {
    await routerToNameWithQuery('Terminal', {
        type: 'container',
        container: row.name,
        containerName: row.name,
        name: row.name,
        operateNode: currentNode.value,
    });
};

const openLog = async (row: any) => {
    await router.push({
        path: '/logs/task',
        query: {
            name: row.name,
        },
    });
};

const toContainer = async (row: any) => {
    await router.push({
        path: '/containers/container',
        query: {
            name: row.name,
        },
    });
};

const openBackups = async (row: any) => {
    await router.push({
        path: '/settings/backupaccount',
        query: {
            appInstallId: String(row.id),
            appKey: row.appKey,
            name: row.name,
        },
    });
};

const openUploads = async (row: any) => {
    await routerToFileWithPath(row.path);
};

const favoriteInstall = async () => {
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const getConfig = async () => {
    try {
        const res = await getAgentSettingByKey('SystemIP');
        if (res.data != '') {
            defaultLink.value = res.data;
            return;
        }
        if (!isMaster.value || currentNodeAddr.value != '127.0.0.1') {
            defaultLink.value = currentNodeAddr.value;
        }
    } catch (error) {}
};

onMounted(() => {
    const path = router.currentRoute.value.path;
    if (path == '/apps/upgrade') {
        activeName.value = i18n.global.t('app.canUpgrade');
        mode.value = 'upgrade';
        searchReq.update = true;
    }
    getConfig();
    search();
});
</script>

<style scoped lang="scss">
@use '../index';

.d-button {
    .el-button + .el-button {
        margin-left: 0;
    }
}
</style>
