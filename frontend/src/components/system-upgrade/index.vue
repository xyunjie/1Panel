<template>
    <div>
        <div class="flex w-full flex-col gap-2 md:flex-row items-center">
            <div class="flex flex-wrap gap-y-2 items-center">
                <span v-if="props.footer">
                    <el-link type="primary" underline="never" @click="toDoc">
                        <span class="font-normal">{{ $t('setting.doc2') }}</span>
                    </el-link>
                    <el-divider direction="vertical" />
                </span>
                <div class="flex flex-wrap items-center">
                    <el-link underline="never" class="version" type="primary">
                        {{ version }}
                    </el-link>
                    <el-tag v-if="version === 'Waiting'" round class="ml-2.5">{{ $t('setting.upgrading') }}</el-tag>
                    <el-link
                        v-if="version !== 'Waiting'"
                        class="ml-2"
                        underline="never"
                        type="primary"
                        @click="upgradeRef.acceptParams()"
                    >
                        {{ $t('commons.button.upgrade') }}
                    </el-link>
                </div>
            </div>
        </div>

        <Upgrade ref="upgradeRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { getSettingInfo } from '@/api/modules/setting';
import Upgrade from '@/components/system-upgrade/upgrade/index.vue';
import { onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { docsUrl } = storeToRefs(globalStore);
const upgradeRef = ref();

const version = ref<string>('');
const props = defineProps({
    footer: { type: Boolean, default: false },
});

const search = async () => {
    const res = await getSettingInfo();
    version.value = res.data.systemVersion;
};

const toDoc = () => {
    window.open(docsUrl.value, '_blank', 'noopener,noreferrer');
};

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
:deep(.el-link__inner) {
    font-weight: 400;
}
.version {
    margin-left: 8px;
    font-size: 14px;
    color: var(--panel-color-primary-light-4);
    text-decoration: none;
    letter-spacing: 0.5px;
    cursor: pointer;
    font-family: auto;
}
</style>
