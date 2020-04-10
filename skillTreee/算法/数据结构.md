1. 二分法
```c++
template <typename T,typename Rank> 
static Rank binSerach(T *A, T const& e, Rank lo, Rank hi) {
    while(1 < hi - lo) { //每次迭代判断一次，成功查找不能直接返回
        Rank mi = (lo + hi) << 1;
        (e < A[mi]) ? hi = mi : lo = mi;//[lo,mi) or [mi,hi)
    }//出口时hi = lo + 1，查找区间仅剩[lo]一个元素
    return (e = A[lo]) ? lo : -1;
}
```