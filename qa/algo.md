# 排序算法
## 快速排序
基本思想是：通过一趟排序将要排序的数据分割成独立的两部分，其中一部分的所有数据都比另外一部分的所有数据都要小， 分割点称为轴心，然后再按此方法对这两部分数据分别进行快速排序，整个排序过程可以递归进行，以此达到整个数据变成有序序列。

```c++
void quick_sort(vector<int>& nums, int l, int r) {
    if (r - l < 2) {
        return;
    }
    // 在[l, r - 1]区间中构造轴心；（算法核心）
    int pivot = partition(nums, l, r - 1);
    //分别对两个区间进行快排
    quick_sort(nums, l, pivot);
    quick_sort(nums, pivot + 1, h);
}

void partition(vector<int>& nums, int l, int r) {
    int first = l;
    int last = r - 1;
    //任选一个元素与首元素交换
    swap(nums[l], nums[l + rand() % (r - l)]);
    key = nums[first];
    while(first < last) {
        while(first < last && nums[last] >= key {
            last--;
        }
        nums[first] = nums[last];
        while(first < last && nums[fist] <= key) {
            fist++;
        }
        nums[last] = nums[fist];
    }
    // 最终first == last； （轴心）
    nums[first] = key;
    return first;
}
```

由于l, r的移动方向相反， 故原处于左端较大的元素将按颠倒的次序转移至右端， 因此**快排并不稳定**。

复杂度：
- 最坏：O(n平方)
- 平均: O(nlog(n))
- 最好: O(nlog(n))

## 归并排序

归并排序可以理解为反复调用二路归并算法得到的。二路归并，就是将两个有序序列合并成一个有序序列。

二路归并属于迭代式算法，每步迭代中，只需要比较两个序列的首元素，将小者取出并追加到输出向量的后端。直到一个序列为空；再将剩余的序列追加到输出向量的后端。

```c++
void merge_sort(vector<int>& nums, int l, int r, vector<int>& temp) {
    if(r - l < 2) {
        return;
    }
    int m = l + (r - l) / 2;
    merge_sort(nums, l, m, temp);
    merge_sort(nums, m, r, temp);

    merge(nums, l, m, r, temp);
}

void merge(vector<int>& nums, int l, int m, int r, vector<int>& temp) {
    int i = l;
    //左区间[l, m), 右区间[m,r)
    //直到左区间和右区间为空
    while(l < m || m < r) {
        //如果右区间为空（左区间必不为空） 或者  左区间不为空，且左端点小于右区间左端点
        if(m >= r || (l < m && nums[l] <= nums[m]) {
            temp[i++] = nums[l++];
        }else {
            temp[i++] = nums[m++]; 
        }
    }
    for(i = l; i < r; i ++) {
        nums[i] = temp[i];
    }
}
```

- 复杂度： O(nlog(n))

- 稳定性： 稳定; 合并过程中我们可以保证如果两个当前元素相等时，我们把处在前面的序列的元素保存在结果序列的前面，这样就保证了稳定性。

## 选择排序

选择排序将序列分为有序的前缀和无序的后缀，且要求后缀不小于后缀，如此，每次只需要从后缀中选择最小者，并作为最大元素转移至前缀中，即可使有序部分不断扩大。

```c++
void selection_sort(vector<int> &nums, int n) {
    int mid;
    for(int i = 0; i < n - 1; i++) {
        mid = i;
        //选取后缀中最小值
        for(int j = i + 1; j n; j++) {
            if(nums[j] < nums[mid]) {
                mid = j;
            }
        }
        //后缀最小值与后缀第一个交换
        swap(nums[mid], nums[i]);
    }
}
```

- 复杂度：o(n平方);(借助高级数据结构可实现O(nlog(n)), 例如: 二叉堆)
- 稳定性：不稳定(例如: 序列5 8 5 2 9)

## 冒泡排序

比较两个相邻的元素，将值大的元素交换到右边；一次遍历后，最大值必在最后；下次遍历(n-1) 即可

```c++
void bubble_sort(vector<int>&nums, int n) {
    bool swapped;
    for (int i = 1; i < n; ++i) {
        swapped = false;
        for (int j = 1; j < n - i + 1; ++j) {
        if (nums[j] < nums[j-1]) {
            swap(nums[j], nums[j-1]);
            swapped = true;
        }
    }
    if (!swapped) {
    break;
}    
```

- 复杂度： O(n平方)

- 稳定性： 稳定;

# todo 
桶排序（top k频次），堆排序（top k 大小）

# 如何在有限的内存限制下实现数十亿级手机号码去重（bitmap）

https://www.jianshu.com/p/b39eb55d4670?utm_campaign=maleskine&utm_content=note&utm_medium=seo_notes&utm_source=recommendation