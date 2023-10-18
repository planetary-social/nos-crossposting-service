import axios, {AxiosResponse} from 'axios';
import {CurrentUser} from "@/dto/CurrentUser";
import {Mutation, State} from '@/store';
import {PublicKeys} from "@/dto/PublicKeys";
import {AddPublicKeyRequest} from "@/dto/AddPublicKeyRequest";
import {Store} from "vuex";

export class APIService {

    private readonly axios = axios.create();

    constructor(private store: Store<State>) {
    }

    currentUser(): Promise<AxiosResponse<CurrentUser>> {
        const url = `/api/current-user`;
        return this.axios.get<CurrentUser>(url);
    }

    publicKeys(): Promise<AxiosResponse<PublicKeys>> {
        const url = `/api/public-keys`;
        return this.axios.get<PublicKeys>(url);
    }

    addPublicKey(req: AddPublicKeyRequest): Promise<AxiosResponse<void>> {
        const url = `/api/public-keys`;
        return this.axios.post<void>(url, req);
    }

    refreshCurrentUser(): Promise<CurrentUser> {
        return new Promise((resolve, reject) => {
            this.currentUser()
                .then(
                    response => {
                        this.store.commit(Mutation.SetUser, response.data.user);
                        resolve(response.data);
                    },
                    error => {
                        reject(error);
                    },
                );
        });
    }
}