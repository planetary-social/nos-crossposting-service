import axios, {AxiosResponse} from 'axios';
import {CurrentUser} from "@/dto/CurrentUser";
import {Mutation, State} from '@/store';
import {PublicKeys} from "@/dto/PublicKeys";
import {AddPublicKeyRequest} from "@/dto/AddPublicKeyRequest";
import {Store} from "vuex";
import {PublicKey} from "@/dto/PublicKey";

export class APIService {

    private readonly axios = axios.create();

    constructor(private store: Store<State>) {
    }

    private currentUser(): Promise<AxiosResponse<CurrentUser>> {
        const url = `/api/current-user`;
        return this.axios.get<CurrentUser>(url);
    }

    private logout(): Promise<AxiosResponse<void>> {
        const url = `/api/current-user`;
        return this.axios.delete<void>(url);
    }

    publicKeys(): Promise<AxiosResponse<PublicKeys>> {
        const url = `/api/current-user/public-keys`;
        return this.axios.get<PublicKeys>(url);
    }

    addPublicKey(req: AddPublicKeyRequest): Promise<AxiosResponse<void>> {
        const url = `/api/current-user/public-keys`;
        return this.axios.post<void>(url, req);
    }

    deletePublicKey(publicKey: PublicKey): Promise<AxiosResponse<void>> {
        const url = `/api/current-user/public-keys/${publicKey.npub}`;
        return this.axios.delete<void>(url);
    }

    logoutCurrentUser(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.logout()
                .then(
                    () => {
                        this.store.commit(Mutation.SetUser, null);
                        resolve();
                    },
                    error => {
                        reject(error);
                    },
                );
        });
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