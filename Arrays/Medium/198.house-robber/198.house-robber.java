class Solution {
    public int rob(int[] nums) {
        if(nums.length<2){
            return nums.length==1?nums[0]:Math.max(nums[0], nums[1]);
        }
        int []dp=new int[nums.length+1];
        dp[1]=nums[0];
        dp[2]=nums[1];
        for(int i=2;i<nums.length;i++){
            dp[i+1]=Math.max(dp[i-2], dp[i-1])+nums[i];
        }
        // System.out.println(Arrays.toString(dp));
        return Math.max(dp[dp.length-1], dp[dp.length-2]);
    }
}