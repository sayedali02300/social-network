import { reactive, ref, watch } from 'vue'
import { API_BASE_URL } from '@/utils/helpers'
import { API_ROUTES, apiURL } from '@/api/api'
import { getUserFollowers, type ConnectionUser } from '@/api/followers';

export function usePostForm() {
    const isSubmitting = ref(false);
    const submitError = ref('');
    const followers = ref<ConnectionUser[]>([]);
    const isLoadingFollowers = ref(false);

    const form = reactive({
        title: '',
        body: '',
        privacy: 'public' as 'public' | 'almost_private' | 'private',
        image: null as File | null,
        selectedFollowers: [] as Array<string>,
    });

    const errors = reactive({
        title: '',
        body: '',
        image: '',
        selectedFollowers: '',
    });

    watch(()=> form.privacy, async(newPrivacy)=>{
        if(newPrivacy === 'private' && followers.value.length === 0){
            isLoadingFollowers.value = true;
            try{
                const meRes = await fetch(apiURL(API_ROUTES.USERS_ME), {method: 'GET', credentials: 'include'});
                const me = await meRes.json();

                followers.value = await getUserFollowers(me.id);
            } catch (err){
                console.error("Error fetching followers:", err);
                errors.selectedFollowers = 'Failed to load followers. Please try again.';
            } finally{
                isLoadingFollowers.value = false;
            }
        }
    });

    const validateTitle = () => {
        const title = form.title.trim();
        if (!title) {
            errors.title = "Title is required";
        } else if (title.length < 3 || title.length > 60) {
            errors.title = "Title must be between 3 to 60 characters";
        } else {
            errors.title = "";
        }
    };

    const validateBody = () => {
        const body = form.body.trim();
        if (!body) {
            errors.body = "Body is required";
        } else if (body.length < 3 || body.length > 5000) {
            errors.body = "Body must be between 3 to 5000 characters";
        } else {
            errors.body = "";
        }
    };

    const handleFileUpload = (event: Event) => {
        const target = event.target as HTMLInputElement;
        if (!target.files || target.files.length === 0) {
            form.image = null;
            return;
        }
        const file = target.files.item(0);
        if (!file) return;

        const validTypes = ["image/jpeg", "image/png", "image/gif"];
        if (!validTypes.includes(file.type)) {
            errors.image = "Invalid file type. Only JPEG, PNG, and GIF allowed.";
            form.image = null;
        } else if (file.size > 10 * 1024 * 1024) {
            errors.image = "File size must be under 10MB.";
            form.image = null;
        } else {
            errors.image = "";
            form.image = file;
        }
    };

    const removeImage = () => {
        form.image = null;
        errors.image = '';
    }


    const submitPost = async (onSuccess: () => void) => {
        validateTitle();
        validateBody();

        const hasErrors = Object.values(errors).some(error => error !== "");
        if (hasErrors) return;

        isSubmitting.value = true;
        submitError.value = '';

        try {
            const formData = new FormData();
            formData.append("title", form.title);
            formData.append("body", form.body);
            formData.append("privacy", form.privacy);
            if (form.image) {
                formData.append("image", form.image);
            }

            if (form.privacy === 'private' && form.selectedFollowers.length > 0) {
                formData.append("allowed_followers", JSON.stringify(form.selectedFollowers));
            }

            const response = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}`, {
                method: "POST",
                body: formData,
                credentials: 'include',
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || "Something went wrong");
            }

            form.title = '';
            form.body = '';
            form.privacy = 'public';
            form.image = null;
            form.selectedFollowers = [];
            errors.title = '';
            errors.body = '';
            errors.image = '';

            onSuccess();

        } catch (err) {
            submitError.value = err instanceof Error ? err.message : 'A network error occurred';
        } finally {
            isSubmitting.value = false;
        }
    };

    return {
        form,
        followers,
        isLoadingFollowers,
        errors,
        isSubmitting,
        submitError,
        validateTitle,
        validateBody,
        handleFileUpload,
        submitPost,
        removeImage
    };
}