import {createStore} from 'vuex'
import {User} from "@/dto/User";

class Notification {
    constructor(
        public style: string,
        public text: string,
    ) {
    }
}

export enum Mutation {
    SetUser = 'setUser',
    PushNotificationError = 'pushNotificationError',
    DismissNotification = 'dismissNotification',
}

export class State {
    user?: User;
    notifications?: Notification[];
}

export default createStore({
    state: {
        user: undefined,
        notifications: [],
    },
    getters: {},
    mutations: {
        [Mutation.SetUser](state: State, user: User): void {
            state.user = user;
        },
        [Mutation.PushNotificationError](state: State, text: string): void {
            state.notifications?.push(new Notification('error', text));
        },
        [Mutation.DismissNotification](state: State, index: number): void {
            state.notifications?.splice(index, 1);
        },
    },
    actions: {},
    modules: {}
})
