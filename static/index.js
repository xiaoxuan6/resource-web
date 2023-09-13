const app = Vue.createApp({
    data() {
        return {
            feeds: [],
            showSEOFlag: true,
            fullscreenLoading: true,
        };
    },
    async created() {
        this.fullscreenLoading = false;
    }
});

app.use(ElementPlus);
app.mount("#app");

function refresh() {
    NProgress.start()
    axios.get('/refresh')
        .then(function (response) {
            Notiflix.Notify.success('刷新成功！');
            NProgress.done()
            setTimeout(function () {
                window.location.reload()
            }, 1000)
        })
        .catch(function (error) {
            Notiflix.Notify.failure(`请求失败: ${error}`);
            NProgress.done()
        })
}