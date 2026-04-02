<template>
    <div>
        <LayoutContent :title="$t('commons.button.set')" v-loading="loading" :divider="true">
            <template #title>
                <div class="flex items-center justify-between">
                    <span>{{ $t('xpack.alert.commonConfig') }}</span>
                    <el-button plain round size="default" @click="onChangeCommon(commonConfig.id)">
                        {{ $t('commons.button.edit') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <el-form
                    @submit.prevent
                    ref="alertFormRef"
                    :label-position="mobile ? 'top' : 'left'"
                    label-width="120px"
                >
                    <el-row>
                        <el-col>
                            <el-form-item :label="$t('xpack.alert.sendTimeRange')" prop="sendTimeRange">
                                {{ sendTimeRange }}
                            </el-form-item>
                            <div v-if="!isMaster">
                                <el-form-item :label="$t('xpack.alert.offline')" prop="isOffline">
                                    <el-switch
                                        @change="onChangeOffline"
                                        v-model="commonConfig.config.isOffline"
                                        active-value="Enable"
                                        inactive-value="Disable"
                                    ></el-switch>
                                    <span class="input-help">{{ $t('xpack.alert.offlineHelper') }}</span>
                                </el-form-item>
                            </div>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>

        <SendTimeRange ref="sendTimeRangeRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, Ref } from 'vue';
import { GlobalStore } from '@/store';
import { ListAlertConfigs, UpdateAlertConfig } from '@/api/modules/alert';
import { ElMessageBox, FormInstance } from 'element-plus';
import SendTimeRange from '@/views/setting/alert/setting/time-range/index.vue';
import i18n from '@/lang';
import { storeToRefs } from 'pinia';
import { MsgSuccess } from '@/utils/message';
import { Alert } from '@/api/interface/alert';

const globalStore = GlobalStore();
const { isMaster } = storeToRefs(globalStore);
const loading = ref(false);

const alertFormRef = ref<FormInstance>();
const sendTimeRangeRef = ref();
const sendTimeRangeValue = ref();
const sendTimeRange = ref();
const isInitialized = ref(false);

const mobile = computed(() => {
    return globalStore.isMobile();
});

const defaultCommonConfig: Alert.CommonAlertConfig = {
    id: undefined,
    type: 'common',
    title: 'xpack.alert.commonConfig',
    status: 'Enable',
    config: {
        alertSendTimeRange:
            i18n.global.t('xpack.alert.noticeAlert') +
            ': ' +
            '08:00:00 - 23:59:59' +
            ' | ' +
            i18n.global.t('xpack.alert.resourceAlert') +
            ': ' +
            '00:00:00 - 23:59:59',
        isOffline: 'Disable',
    },
};

const commonConfig = ref<Alert.CommonAlertConfig>({ ...defaultCommonConfig });

const config = ref<Alert.AlertConfigInfo>({
    id: 0,
    type: '',
    title: '',
    status: '',
    config: '',
});

function parseConfig<T extends object>(raw: any, fallback: T): T {
    try {
        const parsed = JSON.parse(raw.config || '{}');
        return { ...fallback, ...parsed };
    } catch {
        return { ...fallback };
    }
}

function assignConfig<T extends { config: any }>(raw: any, target: Ref<T>, fallback: T) {
    if (raw) {
        target.value = {
            ...(fallback as any),
            id: raw.id,
            type: raw.type,
            title: raw.title,
            status: raw.status,
            config: parseConfig(raw, fallback.config),
        };
    } else {
        target.value = { ...fallback };
    }
}

const search = async () => {
    loading.value = true;
    try {
        const res = await ListAlertConfigs();
        const commonFound = res.data.find((s: any) => s.type === 'common');
        assignConfig(commonFound, commonConfig, defaultCommonConfig);
        sendTimeRangeValue.value = commonConfig.value.config.alertSendTimeRange;
        const noticeTimeRange = sendTimeRangeValue.value?.noticeAlert?.sendTimeRange || '08:00:00 - 23:59:59';
        const resourceTimeRange = sendTimeRangeValue.value?.resourceAlert?.sendTimeRange || '00:00:00 - 23:59:59';
        sendTimeRange.value =
            i18n.global.t('xpack.alert.noticeAlert') +
            ': ' +
            noticeTimeRange +
            ' | ' +
            i18n.global.t('xpack.alert.resourceAlert') +
            ': ' +
            resourceTimeRange;
        isInitialized.value = true;
    } finally {
        loading.value = false;
    }
};

const onChangeCommon = (id: any) => {
    sendTimeRangeRef.value.acceptParams({
        id: id,
        sendTimeRange: sendTimeRangeValue.value,
        isOffline: commonConfig.value.config.isOffline,
    });
};

const onChangeOffline = async () => {
    if (!isInitialized.value) return;
    if (!isMaster.value && commonConfig.value.config.isOffline != '') {
        const title =
            commonConfig.value.config.isOffline == 'Enable'
                ? i18n.global.t('xpack.alert.offlineOff')
                : i18n.global.t('xpack.alert.offlineClose');
        const content =
            commonConfig.value.config.isOffline == 'Enable'
                ? i18n.global.t('xpack.alert.offlineOffHelper')
                : i18n.global.t('xpack.alert.offlineCloseHelper');
        ElMessageBox.confirm(content, title, {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        })
            .then(async () => {
                loading.value = true;
                try {
                    config.value.id = commonConfig.value.id;
                    config.value.type = 'common';
                    config.value.title = 'xpack.alert.commonConfig';
                    config.value.status = 'Enable';
                    config.value.config = JSON.stringify(commonConfig.value.config);
                    await UpdateAlertConfig(config.value);
                    loading.value = false;
                    await search();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                } catch {
                    loading.value = false;
                }
            })
            .catch(() => {
                commonConfig.value.config.isOffline =
                    commonConfig.value.config.isOffline == 'Enable' ? 'Disable' : 'Enable';
            });
    }
};

onMounted(async () => {
    await search();
});
</script>
