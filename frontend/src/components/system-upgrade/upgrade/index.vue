<template>
    <DrawerPro v-model="drawerVisible" :header="$t('commons.button.upgrade')" @close="handleClose" size="small">
        <el-form ref="formRef" :model="form" label-width="120px" label-position="top">
            <el-form-item :label="$t('setting.upgradeFilePath')" prop="filePath" :rules="[{ required: true }]">
                <el-input v-model="form.filePath" :placeholder="$t('setting.upgradeFilePathHelper')" />
            </el-form-item>
            <el-form-item :label="$t('setting.upgradeVersion')" prop="version" :rules="[{ required: true }]">
                <el-input v-model="form.version" placeholder="v2.x.x" />
            </el-form-item>
            <el-form-item :label="$t('setting.upgradeChecksum')" prop="checksum">
                <el-input v-model="form.checksum" :placeholder="$t('setting.upgradeChecksumHelper')" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :loading="loading" @click="onUpgrade">
                    {{ $t('setting.upgradeNow') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { upgradeByFile } from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ref, reactive } from 'vue';
import { GlobalStore } from '@/store';
import { ElMessageBox } from 'element-plus';

const globalStore = GlobalStore();
const drawerVisible = ref(false);
const loading = ref(false);
const formRef = ref();

const form = reactive({ filePath: '', version: '', checksum: '' });

const emit = defineEmits(['search']);

const acceptParams = (): void => {
    form.filePath = '';
    form.version = '';
    form.checksum = '';
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
};

const onUpgrade = async () => {
    await formRef.value?.validate();
    ElMessageBox.confirm(i18n.global.t('setting.upgradeHelper', i18n.global.t('commons.button.upgrade')), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        try {
            await upgradeByFile(form.filePath, form.version, form.checksum);
            globalStore.isLoading = true;
            globalStore.isOnRestart = true;
            drawerVisible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
        } finally {
            loading.value = false;
        }
    });
};

defineExpose({ acceptParams });
</script>
