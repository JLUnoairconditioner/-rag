import { useWebSocket } from '@vueuse/core';
import { localStg } from '@/utils/storage';

export const useChatStore = defineStore(SetupStoreId.Chat, () => {
  const conversationId = ref<string>('');
  const input = ref<Api.Chat.Input>({ message: '' });

  const ragMode = ref<'strict' | 'flexible'>(localStg.get('chatRagMode') ?? 'strict');
  watch(ragMode, val => localStg.set('chatRagMode', val));

  const list = ref<Api.Chat.Message[]>([]);

  const store = useAuthStore();

  const {
    status: wsStatus,
    data: wsData,
    send: wsSend,
    open: wsOpen,
    close: wsClose
  } = useWebSocket(`/proxy-ws/chat/${store.token}`, {
    autoReconnect: true
  });

  const scrollToBottom = ref<null | (() => void)>(null);

  return {
    input,
    conversationId,
    list,
    ragMode,
    wsStatus,
    wsData,
    wsSend,
    wsOpen,
    wsClose,
    scrollToBottom
  };
});
