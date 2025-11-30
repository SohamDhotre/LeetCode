/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode() {}
 *     ListNode(int val) { this.val = val; }
 *     ListNode(int val, ListNode next) { this.val = val; this.next = next; }
 * }
 */
class Solution {
    public ListNode addTwoNumbers(ListNode l1, ListNode l2) {
        ListNode dummy=new ListNode(0);
        int carry=0;
        ListNode list1=l1, list2=l2, list3=dummy;
        
        while(list1!=null || list2!=null || carry!=0){
            // System.out.println("list1: "+(list1!=null?list1.val:"null")+", list2: "+(list2!=null?list2.val:"null"));
            int sum=(list1!=null?list1.val:0)+(list2!=null?list2.val:0)+carry;
            list3.next = new ListNode(sum % 10);
            list3 = list3.next;
            carry=sum/10;
            // print(dummy.next);
            if(list1!=null) list1=list1.next;
            if(list2!=null) list2=list2.next;
            // System.out.println("sum: "+sum+", carry: "+carry+", list1: "
            // +(list1!=null?list1.val:"null")+", list2: "+(list2!=null?list2.val:"null"));
        }
        
        // print(dummy.next);
        return dummy.next;
    }

    void print(ListNode head){
        ListNode cur=head;
        while(cur!=null){
            System.out.print(cur.val+"->");
            cur=cur.next;
        }
    }

    // ListNode reverse(ListNode head){
    //     ListNode, prev=null, cur=head;
    //     while(cur!=null){
    //         ListNode temp=cur.next;
    //         cur.next=prev;
    //         prev=cur;
    //         cur=temp;
    //     }
    //     return prev;
    // }
}