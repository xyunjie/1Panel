<template>
    <div>
        <docker-status v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }" :title="$t('app.app', 2)">
            <template #search>
                <Tags @change="changeTag" />
            </template>
            <template #leftToolBar>
                <el-button @click="syncLocal" type="primary" plain :disabled="syncing">
                    {{ $t('app.syncLocalApp') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-checkbox class="!mr-2.5" v-model="req.showCurrentArch" @change="search(req)">
                    {{ $t('app.showCurrentArch') }}
                </el-checkbox>
                <TableSearch @search="searchByName()" v-model:searchName="req.name" />
            </template>
            <template #main>
                <div>
                    <MainDiv :heightDiff="300">
                        <el-alert
                            type="info"
                            title="应用商店已切换为本地离线目录，仅显示本地可安装应用。"
                            :closable="false"
                        />
                        <el-row :gutter="5" v-if="apps.length > 0">
                            <el-col
                                class="app-col-12"
                                v-for="(app, index) in apps"
                                :key="index"
                                :xs="24"
                                :sm="12"
                                :md="8"
                                :lg="8"
                                :xl="6"
                            >
                                <AppCard :app="app" @open-install="openInstall" @open-detail="openDetail" />
                            </el-col>
                        </el-row>
                        <NoApp v-if="noApp" />
                    </MainDiv>
                    <div class="page-button">
                        <fu-table-pagination
                            v-model:current-page="paginationConfig.currentPage"
                            v-model:page-size="paginationConfig.pageSize"
                            v-bind="paginationConfig"
                            @change="search(req)"
                            :page-sizes="[30, 60, 90]"
                            :layout="mobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
                        />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
    <Install ref="installRef" />
    <Detail ref="detailRef" />
    <TaskLog ref="taskLogRef" @close="refresh" />
</template>

<script lang="ts" setup>
import type { App } from '@/api/interface/app';
import { computed, onMounted, reactive, ref } from 'vue';
import { searchApp, syncLocalApp } from '@/api/modules/app';
import Install from '../detail/install/index.vue';
import router from '@/routers';
import { newUUID } from '@/utils/util';
import Detail from '../detail/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import bus from '@/global/bus';
import Tags from '@/views/app-store/components/tag.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import NoApp from '@/views/app-store/apps/no-app/index.vue';
import AppCard from '@/views/app-store/apps/app/index.vue';
import MainDiv from '@/components/main-div/index.vue';
import { jumpToInstall } from '@/utils/app';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { globalStore } = useGlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const paginationConfig = reactive({
    cacheSizeKey: 'app-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('app-page-size')) || 60,
    total: 0,
});

const req = reactive({
    name: '',
    tags: [],
    page: 1,
    pageSize: 60,
    resource: 'local',
    showCurrentArch: false,
});

const apps = ref<App.AppItem[]>([]);
const loading = ref(false);
const syncing = ref(false);
const installRef = ref();
const installKey = ref('');
const detailRef = ref();
const taskLogRef = ref();
const isActive = ref(false);
const isExist = ref(false);
const noApp = ref(false);

const refresh = () => {
    search(req);
};

const search = async (params: App.AppReq) => {
    loading.value = true;
    params.pageSize = paginationConfig.pageSize;
    params.page = paginationConfig.currentPage;
    localStorage.setItem('app-page-size', params.pageSize + '');

    const res = await searchApp({
        page: params.page,
        pageSize: params.pageSize,
        tags: params.tags,
        name: params.name,
        resource: 'local',
        showCurrentArch: params.showCurrentArch,
    });
    apps.value = res.data.items || [];
    paginationConfig.total = res.data.total || 0;
    noApp.value = apps.value.length === 0;
    loading.value = false;
};

const openInstall = (app: App.App) => {
    if (!jumpToInstall(app.type, app.key)) {
        installRef.value.acceptParams({ app });
    }
};

const openDetail = (key: string) => {
    detailRef.value.acceptParams(key, 'install');
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const syncLocal = () => {
    const taskID = newUUID();
    syncing.value = true;
    syncLocalApp({ taskID })
        .then(() => {
            openTaskLog(taskID);
            search(req);
        })
        .finally(() => {
            syncing.value = false;
        });
};

const changeTag = (key: string) => {
    req.tags = [];
    if (key !== 'all') {
        req.tags = [key];
    }
    search(req);
};

const searchByName = () => {
    search(req);
};

onMounted(() => {
    bus.on('refreshApp', () => {
        search(req);
    });
    if (router.currentRoute.value.query.install) {
        installKey.value = String(router.currentRoute.value.query.install);
        installRef.value.acceptParams({ app: { key: installKey.value } });
    }
    search(req);
});
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) and (max-width: 1200px) {
    .app-col-12 {
        max-width: 50%;
        flex: 0 0 50%;
    }
}

.page-button {
    float: right;
    margin-bottom: 10px;
    margin-top: 10px;
}
</style>
