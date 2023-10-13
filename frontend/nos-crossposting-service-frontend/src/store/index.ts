import {createStore} from 'vuex'
import {User} from "@/dto/User";

export enum Mutation {
    SetUser = 'setUser',
}

export class State {
    user?: User;
}

export default createStore({
    state: {
        user: undefined,
    },
    getters: {},
    mutations: {
        [Mutation.SetUser](state: State, user: User): void {
            state.user = user;
        },
    },
    actions: {},
    modules: {}
})
